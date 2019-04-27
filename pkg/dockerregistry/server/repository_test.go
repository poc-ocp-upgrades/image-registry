package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
	"github.com/docker/distribution"
	dockercfg "github.com/docker/distribution/configuration"
	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/inmemory"
	"github.com/docker/libtrust"
	"github.com/opencontainers/go-digest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	clientgotesting "k8s.io/client-go/testing"
	imageapiv1 "github.com/openshift/api/image/v1"
	imageapi "github.com/openshift/image-registry/pkg/origin-common/image/apis/image"
	"github.com/openshift/image-registry/pkg/origin-common/util"
	registryclient "github.com/openshift/image-registry/pkg/dockerregistry/server/client"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/configuration"
	"github.com/openshift/image-registry/pkg/testutil"
)

const (
	testImageLayerCount = 2
)

func TestRepositoryBlobStat(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	backgroundCtx := context.Background()
	backgroundCtx = testutil.WithTestLogger(backgroundCtx, t)
	req, err := http.NewRequest("GET", "https://localhost:5000/nm/is", nil)
	if err != nil {
		t.Fatal(err)
	}
	backgroundCtx = dcontext.WithRequest(backgroundCtx, req)
	backgroundCtx = withAppMiddleware(backgroundCtx, &fakeAccessControllerMiddleware{t: t})
	cfg := &configuration.Configuration{Server: &configuration.Server{Addr: "localhost:5000"}}
	if err := configuration.InitExtraConfig(&dockercfg.Configuration{}, cfg); err != nil {
		t.Fatal(err)
	}
	driver := inmemory.New()
	testImages, err := populateTestStorage(backgroundCtx, t, driver, true, 1, map[string]int{"nm/is:latest": 1, "nm/repo:missing-layer-links": 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	testImages, err = populateTestStorage(backgroundCtx, t, driver, false, 1, map[string]int{"nm/unmanaged:missing-layer-links": 1}, testImages)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"nm/repo:missing-layer-links", "nm/unmanaged:missing-layer-links"} {
		repoName := strings.Split(name, ":")[0]
		for _, layer := range testImages[name][0].DockerImageLayers {
			dgst := digest.Digest(layer.Name)
			alg, hex := dgst.Algorithm(), dgst.Hex()
			err := driver.Delete(backgroundCtx, fmt.Sprintf("/docker/registry/v2/repositories/%s/_layers/%s/%s", repoName, alg, hex))
			if err != nil {
				t.Fatalf("failed to delete layer link %q from repository %q: %v", layer.Name, repoName, err)
			}
		}
	}
	etcdOnlyImages := map[string]*imageapiv1.Image{}
	for _, d := range []struct {
		name	string
		managed	bool
	}{{"nm/is", true}, {"registry.org:5000/user/app", false}} {
		img, err := testutil.NewImageForManifest(d.name, testutil.SampleImageManifestSchema1, "", d.managed)
		if err != nil {
			t.Fatal(err)
		}
		etcdOnlyImages[d.name] = img
	}
	for _, tc := range []struct {
		name			string
		stat			string
		images			[]imageapiv1.Image
		imageStreams		[]imageapiv1.ImageStream
		skipAuth		bool
		deferredErrors		deferredErrors
		expectedDescriptor	distribution.Descriptor
		expectedError		error
		expectedActions		[]clientAction
	}{{name: "local stat", stat: "nm/is@" + testImages["nm/is:latest"][0].DockerImageLayers[0].Name, imageStreams: []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "is"}}}, expectedDescriptor: testNewDescriptorForLayer(testImages["nm/is:latest"][0].DockerImageLayers[0])}, {name: "blob only tagged in image stream", stat: "nm/repo@" + testImages["nm/repo:missing-layer-links"][0].DockerImageLayers[1].Name, images: []imageapiv1.Image{*testImages["nm/repo:missing-layer-links"][0]}, imageStreams: []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "repo"}, Status: imageapiv1.ImageStreamStatus{Tags: []imageapiv1.NamedTagEventList{{Tag: "latest", Items: []imageapiv1.TagEvent{{Image: testImages["nm/repo:missing-layer-links"][0].Name}}}}}}}, expectedDescriptor: testNewDescriptorForLayer(testImages["nm/repo:missing-layer-links"][0].DockerImageLayers[1]), expectedActions: []clientAction{{"get", "imagestreams/layers"}}}, {name: "blob referenced only by unmanaged image with pullthrough", stat: "nm/unmanaged@" + testImages["nm/unmanaged:missing-layer-links"][0].DockerImageLayers[1].Name, images: []imageapiv1.Image{*testImages["nm/unmanaged:missing-layer-links"][0]}, imageStreams: []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "unmanaged"}, Status: imageapiv1.ImageStreamStatus{Tags: []imageapiv1.NamedTagEventList{{Tag: "latest", Items: []imageapiv1.TagEvent{{Image: testImages["nm/unmanaged:missing-layer-links"][0].Name}}}}}}}, expectedDescriptor: testNewDescriptorForLayer(testImages["nm/unmanaged:missing-layer-links"][0].DockerImageLayers[1]), expectedActions: []clientAction{{"get", "imagestreams/layers"}}}, {name: "blob not referenced", stat: "nm/unmanaged@" + testImages["nm/unmanaged:missing-layer-links"][0].DockerImageLayers[1].Name, imageStreams: []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "unmanaged"}, Status: imageapiv1.ImageStreamStatus{Tags: []imageapiv1.NamedTagEventList{{Tag: "latest", Items: []imageapiv1.TagEvent{{Image: testImages["nm/unmanaged:missing-layer-links"][0].Name}}}}}}}, expectedError: distribution.ErrBlobUnknown, expectedActions: []clientAction{{"get", "imagestreams/layers"}, {"get", "imagestreams"}, {"get", "imagestreams/secrets"}}}, {name: "layer link present while image stream not found", stat: "nm/is@" + testImages["nm/is:latest"][0].DockerImageLayers[0].Name, images: []imageapiv1.Image{*testImages["nm/is:latest"][0]}, expectedDescriptor: testNewDescriptorForLayer(testImages["nm/is:latest"][0].DockerImageLayers[0])}, {name: "blob not stored locally but referred in image stream", stat: "nm/is@" + etcdOnlyImages["nm/is"].DockerImageLayers[1].Name, images: []imageapiv1.Image{*etcdOnlyImages["nm/is"]}, imageStreams: []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "is"}, Status: imageapiv1.ImageStreamStatus{Tags: []imageapiv1.NamedTagEventList{{Tag: "latest", Items: []imageapiv1.TagEvent{{Image: etcdOnlyImages["nm/is"].Name}}}}}}}, expectedError: distribution.ErrBlobUnknown, expectedActions: []clientAction{{"get", "imagestreams"}, {"get", "imagestreams/secrets"}}}, {name: "blob does not exist", stat: "nm/repo@" + etcdOnlyImages["nm/is"].DockerImageLayers[0].Name, images: []imageapiv1.Image{*testImages["nm/is:latest"][0]}, imageStreams: []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "repo"}, Status: imageapiv1.ImageStreamStatus{Tags: []imageapiv1.NamedTagEventList{{Tag: "latest", Items: []imageapiv1.TagEvent{{Image: testImages["nm/is:latest"][0].Name}}}}}}}, expectedError: distribution.ErrBlobUnknown, expectedActions: []clientAction{{"get", "imagestreams"}, {"get", "imagestreams/secrets"}}}, {name: "auth not performed", stat: "nm/is@" + testImages["nm/is:latest"][0].DockerImageLayers[0].Name, imageStreams: []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "is"}}}, skipAuth: true, expectedError: fmt.Errorf("openshift.auth.completed missing from context")}, {name: "deferred error", stat: "nm/is@" + testImages["nm/is:latest"][0].DockerImageLayers[0].Name, imageStreams: []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "is"}}}, deferredErrors: deferredErrors{"nm/is": ErrOpenShiftAccessDenied}, expectedError: ErrOpenShiftAccessDenied}} {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := reference.Parse(tc.stat)
			if err != nil {
				t.Fatalf("failed to parse blob reference %q: %v", tc.stat, err)
			}
			canonical, ok := ref.(reference.Canonical)
			if !ok {
				t.Fatalf("not a canonical reference %q", ref.String())
			}
			ctx := backgroundCtx
			if !tc.skipAuth {
				ctx = withAuthPerformed(ctx)
			}
			if tc.deferredErrors != nil {
				ctx = withDeferredErrors(ctx, tc.deferredErrors)
			}
			fos, imageClient := testutil.NewFakeOpenShiftWithClient(ctx)
			for _, is := range tc.imageStreams {
				_, err = fos.CreateImageStream(is.Namespace, &is)
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, image := range tc.images {
				_, err = fos.CreateImage(&image)
				if err != nil {
					t.Fatal(err)
				}
			}
			reg, err := newTestRegistry(ctx, registryclient.NewFakeRegistryAPIClient(nil, imageClient), driver, cfg.Cache.BlobRepositoryTTL, true)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			repo, err := reg.Repository(ctx, canonical)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			desc, err := repo.Blobs(ctx).Stat(ctx, canonical.Digest())
			if err != nil && tc.expectedError == nil {
				t.Fatalf("got unexpected stat error: %v", err)
			}
			if err == nil && tc.expectedError != nil {
				t.Fatalf("got unexpected non-error")
			}
			if !reflect.DeepEqual(err, tc.expectedError) {
				t.Fatalf("got unexpected error: %s", diff.ObjectGoPrintDiff(err, tc.expectedError))
			}
			if tc.expectedError == nil && !reflect.DeepEqual(desc, tc.expectedDescriptor) {
				t.Fatalf("got unexpected descriptor: %s", diff.ObjectGoPrintDiff(desc, tc.expectedDescriptor))
			}
			compareActions(t, tc.name, imageClient.Actions(), tc.expectedActions)
		})
	}
}
func TestRepositoryBlobStatCacheEviction(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	const blobRepoCacheTTL = time.Millisecond * 500
	ctx := context.Background()
	ctx = testutil.WithTestLogger(ctx, t)
	ctx = withAuthPerformed(ctx)
	driver := inmemory.New()
	testImages, err := populateTestStorage(ctx, t, driver, true, 1, map[string]int{"nm/is:latest": 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	testImage := testImages["nm/is:latest"][0]
	blob1Desc := testNewDescriptorForLayer(testImage.DockerImageLayers[0])
	blob1Dgst := blob1Desc.Digest
	blob2Desc := testNewDescriptorForLayer(testImage.DockerImageLayers[1])
	blob2Dgst := blob2Desc.Digest
	alg, hex := blob2Dgst.Algorithm(), blob2Dgst.Hex()
	err = driver.Delete(ctx, fmt.Sprintf("/docker/registry/v2/repositories/%s/_layers/%s/%s", "nm/is", alg, hex))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fos, imageClient := testutil.NewFakeOpenShiftWithClient(ctx)
	testutil.AddImageStream(t, fos, "nm", "is", nil)
	testutil.AddImage(t, fos, testImage, "nm", "is", "latest")
	reg, err := newTestRegistry(ctx, registryclient.NewFakeRegistryAPIClient(nil, imageClient), driver, blobRepoCacheTTL, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ref, err := reference.WithName("nm/is")
	if err != nil {
		t.Errorf("failed to parse blob reference %q: %v", "nm/is", err)
	}
	repo, err := reg.Repository(ctx, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	desc, err := repo.Blobs(ctx).Stat(ctx, blob1Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if !reflect.DeepEqual(desc, blob1Desc) {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob1Desc)
	}
	compareActions(t, "no actions expected", imageClient.Actions(), []clientAction{})
	err = repo.Blobs(ctx).Delete(ctx, blob1Dgst)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	repo, err = reg.Repository(ctx, ref)
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	desc, err = repo.Blobs(ctx).Stat(ctx, blob1Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if !reflect.DeepEqual(desc, blob1Desc) {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob1Desc)
	}
	expectedActions := []clientAction{{"get", "imagestreams/layers"}}
	compareActions(t, "1st roundtrip to etcd", imageClient.Actions(), expectedActions)
	vacuum := storage.NewVacuum(ctx, driver)
	err = vacuum.RemoveBlob(blob1Dgst.String())
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	repo, err = reg.Repository(ctx, ref)
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	desc, err = repo.Blobs(ctx).Stat(ctx, blob2Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if desc.Digest != blob2Desc.Digest || desc.Size != blob2Desc.Size {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob2Desc)
	}
	compareActions(t, "no etcd query", imageClient.Actions(), expectedActions)
	lastStatTimestamp := time.Now()
	repo, err = reg.Repository(ctx, ref)
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	desc, err = repo.Blobs(ctx).Stat(ctx, blob2Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if desc.Digest != blob2Desc.Digest || desc.Size != blob2Desc.Size {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob2Desc)
	}
	compareActions(t, "no roundrip to etcd", imageClient.Actions(), expectedActions)
	t.Logf("sleeping %s while waiting for eviction of blob %q from cache", blobRepoCacheTTL.String(), blob2Dgst.String())
	time.Sleep(blobRepoCacheTTL - time.Since(lastStatTimestamp))
	repo, err = reg.Repository(ctx, ref)
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	desc, err = repo.Blobs(ctx).Stat(ctx, blob2Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if !reflect.DeepEqual(desc, blob2Desc) {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob2Desc)
	}
	expectedActions = append(expectedActions, []clientAction{{"get", "imagestreams/layers"}}...)
	compareActions(t, "2nd roundtrip to etcd", imageClient.Actions(), expectedActions)
	err = vacuum.RemoveBlob(blob2Dgst.String())
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
}

type clientAction struct {
	verb		string
	resource	string
}

func storeTestImage(ctx context.Context, reg distribution.Namespace, imageReference reference.NamedTagged, schemaVersion int, managedByOpenShift bool) (*imageapiv1.Image, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	repo, err := reg.Repository(ctx, imageReference)
	if err != nil {
		return nil, fmt.Errorf("unexpected error getting repo %q: %v", imageReference.Name(), err)
	}
	var (
		m	distribution.Manifest
		m1	schema1.Manifest
	)
	switch schemaVersion {
	case 1:
		m1 = schema1.Manifest{Versioned: manifest.Versioned{SchemaVersion: 1}, Name: imageReference.Name(), Tag: imageReference.Tag()}
	case 2:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported manifest version %d", schemaVersion)
	}
	for i := 0; i < testImageLayerCount; i++ {
		payload, err := testutil.CreateRandomTarFile()
		if err != nil {
			return nil, fmt.Errorf("unexpected error generating test layer file: %v", err)
		}
		wr, err := repo.Blobs(ctx).Create(ctx)
		if err != nil {
			return nil, fmt.Errorf("unexpected error creating test upload: %v", err)
		}
		defer wr.Close()
		n, err := io.Copy(wr, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("unexpected error copying to upload: %v", err)
		}
		dgst := digest.FromBytes(payload)
		if schemaVersion == 1 {
			m1.FSLayers = append(m1.FSLayers, schema1.FSLayer{BlobSum: dgst})
			m1.History = append(m1.History, schema1.History{V1Compatibility: fmt.Sprintf(`{"size":%d}`, n)})
		}
		if _, err := wr.Commit(ctx, distribution.Descriptor{Digest: dgst, MediaType: schema1.MediaTypeManifestLayer}); err != nil {
			return nil, fmt.Errorf("unexpected error finishing upload: %v", err)
		}
	}
	var dgst digest.Digest
	var payload []byte
	if schemaVersion == 1 {
		pk, err := libtrust.GenerateECP256PrivateKey()
		if err != nil {
			return nil, fmt.Errorf("unexpected error generating private key: %v", err)
		}
		m, err = schema1.Sign(&m1, pk)
		if err != nil {
			return nil, fmt.Errorf("error signing manifest: %v", err)
		}
		_, payload, err = m.Payload()
		if err != nil {
			return nil, fmt.Errorf("error getting payload %#v", err)
		}
		dgst = digest.FromBytes(payload)
	}
	image := &imageapi.Image{ObjectMeta: metav1.ObjectMeta{Name: dgst.String()}, DockerImageManifest: string(payload), DockerImageReference: imageReference.Name() + "@" + dgst.String()}
	if managedByOpenShift {
		image.Annotations = map[string]string{imageapi.ManagedByOpenShiftAnnotation: "true"}
	}
	if schemaVersion == 1 {
		signedManifest := m.(*schema1.SignedManifest)
		signatures, err := signedManifest.Signatures()
		if err != nil {
			return nil, err
		}
		image.DockerImageSignatures = append(image.DockerImageSignatures, signatures...)
	}
	if err := util.InternalImageWithMetadata(image); err != nil {
		return nil, err
	}
	newImage, err := testutil.ConvertImage(image)
	if err != nil {
		return nil, fmt.Errorf("failed to convert image from internal to external type: %v", err)
	}
	if err := util.ImageWithMetadata(newImage); err != nil {
		return nil, fmt.Errorf("failed to fill image with metadata: %v", err)
	}
	return newImage, nil
}
func populateTestStorage(ctx context.Context, t *testing.T, driver driver.StorageDriver, setManagedByOpenShift bool, schemaVersion int, repoImages map[string]int, testImages map[string][]*imageapiv1.Image) (map[string][]*imageapiv1.Image, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	reg, err := storage.NewRegistry(ctx, driver)
	if err != nil {
		t.Fatalf("error creating registry: %v", err)
	}
	result := make(map[string][]*imageapiv1.Image)
	for key, value := range testImages {
		images := make([]*imageapiv1.Image, len(value))
		copy(images, value)
		result[key] = images
	}
	for imageReference := range repoImages {
		parsed, err := reference.Parse(imageReference)
		if err != nil {
			t.Fatalf("failed to parse reference %q: %v", imageReference, err)
		}
		namedTagged, ok := parsed.(reference.NamedTagged)
		if !ok {
			t.Fatalf("expected NamedTagged reference, not %T", parsed)
		}
		imageCount := repoImages[imageReference]
		for i := 0; i < imageCount; i++ {
			img, err := storeTestImage(ctx, reg, namedTagged, schemaVersion, setManagedByOpenShift)
			if err != nil {
				t.Fatal(err)
			}
			arr := result[imageReference]
			t.Logf("created image %s@%s image with layers:", namedTagged.Name(), img.Name)
			for _, l := range img.DockerImageLayers {
				t.Logf("  %s of size %d", l.Name, l.LayerSize)
			}
			result[imageReference] = append(arr, img)
		}
	}
	return result, nil
}
func testNewDescriptorForLayer(layer imageapiv1.ImageLayer) distribution.Descriptor {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return distribution.Descriptor{Digest: digest.Digest(layer.Name), MediaType: "application/octet-stream", Size: layer.LayerSize}
}
func compareActions(t *testing.T, testCaseName string, actions []clientgotesting.Action, expectedActions []clientAction) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	t.Helper()
	for i, action := range actions {
		if i >= len(expectedActions) {
			t.Errorf("got unexpected client action: %#+v", action)
			continue
		}
		expected := expectedActions[i]
		parts := strings.Split(expected.resource, "/")
		if !action.Matches(expected.verb, parts[0]) {
			t.Errorf("expected client action %#+v at index %d, got instead: %#+v", expected, i, action)
		}
		if (len(parts) > 1 && action.GetSubresource() != parts[1]) || (len(parts) == 1 && len(action.GetSubresource()) > 0) {
			t.Errorf("expected client action %#+v at index %d, got instead: %#+v", expected, i, action)
		}
	}
	for i := len(actions); i < len(expectedActions); i++ {
		expected := expectedActions[i]
		t.Errorf("expected action %#v did not happen (%#v)", expected, actions)
	}
}

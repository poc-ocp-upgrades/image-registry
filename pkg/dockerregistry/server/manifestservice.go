package server

import (
	"context"
	"fmt"
	"github.com/docker/distribution"
	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/registry/api/errcode"
	regapi "github.com/docker/distribution/registry/api/v2"
	"github.com/opencontainers/go-digest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	imageapiv1 "github.com/openshift/api/image/v1"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/cache"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/manifesthandler"
	"github.com/openshift/image-registry/pkg/imagestream"
	imageapi "github.com/openshift/image-registry/pkg/origin-common/image/apis/image"
)

var _ distribution.ManifestService = &manifestService{}

type manifestService struct {
	manifests		distribution.ManifestService
	blobStore		distribution.BlobStore
	serverAddr		string
	imageStream		imagestream.ImageStream
	cache			cache.RepositoryDigest
	acceptSchema2	bool
}

func (m *manifestService) Exists(ctx context.Context, dgst digest.Digest) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*manifestService).Exists")
	image, err := m.imageStream.GetImageOfImageStream(ctx, dgst)
	if err != nil {
		switch err.Code {
		case imagestream.ErrImageStreamImageNotFoundCode:
			dcontext.GetLogger(ctx).Errorf("manifestService.Exists: image %s is not found in imagestream %s", dgst.String(), m.imageStream.Reference())
			fallthrough
		case imagestream.ErrImageStreamNotFoundCode:
			return false, distribution.ErrBlobUnknown
		}
		return false, err
	}
	return image != nil, nil
}
func (m *manifestService) Get(ctx context.Context, dgst digest.Digest, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*manifestService).Get")
	image, rErr := m.imageStream.GetImageOfImageStream(ctx, dgst)
	if rErr != nil {
		switch rErr.Code {
		case imagestream.ErrImageStreamNotFoundCode, imagestream.ErrImageStreamImageNotFoundCode:
			dcontext.GetLogger(ctx).Errorf("manifestService.Get: unable to get image %s in imagestream %s: %v", dgst.String(), m.imageStream.Reference(), rErr)
			return nil, distribution.ErrManifestUnknownRevision{Name: m.imageStream.Reference(), Revision: dgst}
		case imagestream.ErrImageStreamForbiddenCode:
			dcontext.GetLogger(ctx).Errorf("manifestService.Get: unable to get access to imagestream %s to find image %s: %v", m.imageStream.Reference(), dgst.String(), rErr)
			return nil, distribution.ErrAccessDenied
		}
		return nil, rErr
	}
	ref := m.imageStream.Reference()
	if !imagestream.IsImageManaged(image) {
		ref = fmt.Sprintf("%s/%s", m.serverAddr, ref)
	}
	manifest, err := m.manifests.Get(ctx, dgst, options...)
	if err != nil {
		return nil, err
	}
	RememberLayersOfImage(ctx, m.cache, image, ref)
	return manifest, nil
}
func (m *manifestService) Put(ctx context.Context, manifest distribution.Manifest, options ...distribution.ManifestServiceOption) (digest.Digest, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*manifestService).Put")
	mh, err := manifesthandler.NewManifestHandler(m.serverAddr, m.blobStore, manifest)
	if err != nil {
		return "", regapi.ErrorCodeManifestInvalid.WithDetail(err)
	}
	mediaType, payload, _, err := mh.Payload()
	if err != nil {
		return "", regapi.ErrorCodeManifestInvalid.WithDetail(err)
	}
	if !m.acceptSchema2 && mediaType == schema2.MediaTypeManifest {
		return "", regapi.ErrorCodeManifestInvalid.WithDetail(fmt.Errorf("manifest V2 schema 2 not allowed"))
	}
	if err := mh.Verify(ctx, false); err != nil {
		return "", err
	}
	_, err = m.manifests.Put(ctx, manifest, options...)
	if err != nil {
		return "", err
	}
	config, err := mh.Config(ctx)
	if err != nil {
		return "", err
	}
	dgst, err := mh.Digest()
	if err != nil {
		return "", err
	}
	layerOrder, layers, err := mh.Layers(ctx)
	if err != nil {
		return "", err
	}
	uclient, ok := userClientFrom(ctx)
	if !ok {
		errmsg := "error creating user client to auto provision image stream: user client to master API unavailable"
		dcontext.GetLogger(ctx).Errorf(errmsg)
		return "", errcode.ErrorCodeUnknown.WithDetail(errmsg)
	}
	image := &imageapiv1.Image{ObjectMeta: metav1.ObjectMeta{Name: dgst.String(), Annotations: map[string]string{imageapi.ManagedByOpenShiftAnnotation: "true", imageapi.ImageManifestBlobStoredAnnotation: "true", imageapi.DockerImageLayersOrderAnnotation: layerOrder}}, DockerImageReference: fmt.Sprintf("%s/%s@%s", m.serverAddr, m.imageStream.Reference(), dgst.String()), DockerImageManifest: string(payload), DockerImageManifestMediaType: mediaType, DockerImageConfig: string(config), DockerImageLayers: layers}
	tag := ""
	for _, option := range options {
		if opt, ok := option.(distribution.WithTagOption); ok {
			tag = opt.Tag
			break
		}
	}
	rErr := m.imageStream.CreateImageStreamMapping(ctx, uclient, tag, image)
	if rErr != nil {
		switch rErr.Code {
		case imagestream.ErrImageStreamNotFoundCode:
			dcontext.GetLogger(ctx).Errorf("manifestService.Put: imagestreammapping failed for image %s@%s: %v", m.imageStream.Reference(), image.Name, rErr)
			return "", distribution.ErrManifestUnknownRevision{Name: m.imageStream.Reference(), Revision: dgst}
		case imagestream.ErrImageStreamForbiddenCode:
			dcontext.GetLogger(ctx).Errorf("manifestService.Put: imagestreammapping got access denied for image %s@%s: %v", m.imageStream.Reference(), image.Name, rErr)
			return "", distribution.ErrAccessDenied
		}
		return "", rErr
	}
	return dgst, nil
}
func (m *manifestService) Delete(ctx context.Context, dgst digest.Digest) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*manifestService).Delete")
	_, err := m.imageStream.GetImageOfImageStream(ctx, dgst)
	if err == nil {
		return distribution.ErrUnsupported
	}
	switch err.Code {
	case imagestream.ErrImageStreamNotFoundCode, imagestream.ErrImageStreamImageNotFoundCode:
	case imagestream.ErrImageStreamForbiddenCode:
		dcontext.GetLogger(ctx).Errorf("manifestService.Delete: unable to get access to imagestream %s to find image %s: %v", m.imageStream.Reference(), dgst.String(), err)
		return distribution.ErrAccessDenied
	default:
		return err
	}
	return m.manifests.Delete(ctx, dgst)
}

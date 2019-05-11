package manifesthandler

import (
	"context"
	"errors"
	"github.com/docker/distribution"
	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	imageapiv1 "github.com/openshift/api/image/v1"
	imageapi "github.com/openshift/image-registry/pkg/origin-common/image/apis/image"
)

var (
	errMissingURL		= errors.New("missing URL on layer")
	errUnexpectedURL	= errors.New("unexpected URL on layer")
)

type manifestSchema2Handler struct {
	blobStore		distribution.BlobStore
	manifest		*schema2.DeserializedManifest
	cachedConfig	[]byte
}

var _ ManifestHandler = &manifestSchema2Handler{}

func (h *manifestSchema2Handler) Config(ctx context.Context) ([]byte, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if h.cachedConfig == nil {
		blob, err := h.blobStore.Get(ctx, h.manifest.Config.Digest)
		if err != nil {
			dcontext.GetLogger(ctx).Errorf("failed to get manifest config: %v", err)
			return nil, err
		}
		h.cachedConfig = blob
	}
	return h.cachedConfig, nil
}
func (h *manifestSchema2Handler) Digest() (digest.Digest, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, p, err := h.manifest.Payload()
	if err != nil {
		return "", err
	}
	return digest.FromBytes(p), nil
}
func (h *manifestSchema2Handler) Manifest() distribution.Manifest {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return h.manifest
}
func (h *manifestSchema2Handler) Layers(ctx context.Context) (string, []imageapiv1.ImageLayer, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	layers := make([]imageapiv1.ImageLayer, len(h.manifest.Layers))
	for i, layer := range h.manifest.Layers {
		layers[i].Name = layer.Digest.String()
		layers[i].LayerSize = layer.Size
		layers[i].MediaType = layer.MediaType
	}
	return imageapi.DockerImageLayersOrderAscending, layers, nil
}
func (h *manifestSchema2Handler) Payload() (mediaType string, payload []byte, canonical []byte, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	mt, p, err := h.manifest.Payload()
	return mt, p, p, err
}
func (h *manifestSchema2Handler) verifyLayer(ctx context.Context, fsLayer distribution.Descriptor) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if fsLayer.MediaType == schema2.MediaTypeForeignLayer {
		if len(fsLayer.URLs) == 0 {
			return errMissingURL
		}
		return nil
	}
	if len(fsLayer.URLs) != 0 {
		return errUnexpectedURL
	}
	desc, err := h.blobStore.Stat(ctx, fsLayer.Digest)
	if err != nil {
		return err
	}
	if fsLayer.Size != desc.Size {
		return ErrManifestBlobBadSize{Digest: fsLayer.Digest, ActualSize: desc.Size, SizeInManifest: fsLayer.Size}
	}
	return nil
}
func (h *manifestSchema2Handler) Verify(ctx context.Context, skipDependencyVerification bool) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	var errs distribution.ErrManifestVerification
	if skipDependencyVerification {
		return nil
	}
	for _, fsLayer := range h.manifest.References() {
		if err := h.verifyLayer(ctx, fsLayer); err != nil {
			if err != distribution.ErrBlobUnknown {
				errs = append(errs, err)
				continue
			}
			errs = append(errs, distribution.ErrManifestBlobUnknown{Digest: fsLayer.Digest})
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

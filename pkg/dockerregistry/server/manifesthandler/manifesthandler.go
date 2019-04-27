package manifesthandler

import (
	"context"
	"fmt"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	imageapiv1 "github.com/openshift/api/image/v1"
)

type ManifestHandler interface {
	Config(ctx context.Context) ([]byte, error)
	Digest() (manifestDigest digest.Digest, err error)
	Manifest() distribution.Manifest
	Layers(ctx context.Context) (order string, layers []imageapiv1.ImageLayer, err error)
	Payload() (mediaType string, payload []byte, canonical []byte, err error)
	Verify(ctx context.Context, skipDependencyVerification bool) error
}

func NewManifestHandler(serverAddr string, blobStore distribution.BlobStore, manifest distribution.Manifest) (ManifestHandler, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch t := manifest.(type) {
	case *schema1.SignedManifest:
		return &manifestSchema1Handler{serverAddr: serverAddr, blobStore: blobStore, manifest: t}, nil
	case *schema2.DeserializedManifest:
		return &manifestSchema2Handler{blobStore: blobStore, manifest: t}, nil
	default:
		return nil, fmt.Errorf("unsupported manifest type %T", manifest)
	}
}

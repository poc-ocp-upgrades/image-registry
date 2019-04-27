package wrapped

import (
	"context"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
)

type repository struct {
	repository	distribution.Repository
	wrapper		Wrapper
}

var _ distribution.Repository = &repository{}

func NewRepository(r distribution.Repository, wrapper Wrapper) distribution.Repository {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &repository{repository: r, wrapper: wrapper}
}
func (r *repository) Named() reference.Named {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r.repository.Named()
}
func (r *repository) Blobs(ctx context.Context) distribution.BlobStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	bs := r.repository.Blobs(ctx)
	return NewBlobStore(bs, r.wrapper)
}
func (r *repository) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	ms, err := r.repository.Manifests(ctx, options...)
	if err != nil {
		return ms, err
	}
	return NewManifestService(ms, r.wrapper), nil
}
func (r *repository) Tags(ctx context.Context) distribution.TagService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	ts := r.repository.Tags(ctx)
	return NewTagService(ts, r.wrapper)
}

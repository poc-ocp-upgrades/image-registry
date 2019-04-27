package server

import (
	"context"
	"github.com/docker/distribution"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/wrapped"
)

func newPendingErrorsWrapper(repo *repository) wrapped.Wrapper {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return func(ctx context.Context, funcname string, f func(ctx context.Context) error) error {
		if err := repo.checkPendingErrors(ctx); err != nil {
			return err
		}
		return f(ctx)
	}
}
func newPendingErrorsBlobStore(bs distribution.BlobStore, repo *repository) distribution.BlobStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return wrapped.NewBlobStore(bs, newPendingErrorsWrapper(repo))
}
func newPendingErrorsManifestService(ms distribution.ManifestService, repo *repository) distribution.ManifestService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return wrapped.NewManifestService(ms, newPendingErrorsWrapper(repo))
}
func newPendingErrorsTagService(ts distribution.TagService, repo *repository) distribution.TagService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return wrapped.NewTagService(ts, newPendingErrorsWrapper(repo))
}
func newPendingErrorsBlobDescriptorService(bds distribution.BlobDescriptorService, repo *repository) distribution.BlobDescriptorService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return wrapped.NewBlobDescriptorService(bds, newPendingErrorsWrapper(repo))
}

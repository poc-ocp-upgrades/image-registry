package server

import (
	"context"
	"errors"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
)

type registry struct {
	registry	distribution.Namespace
	enumerator	RepositoryEnumerator
}

var _ distribution.Namespace = &registry{}

func (r *registry) Scope() distribution.Scope {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r.registry.Scope()
}
func (r *registry) Repository(ctx context.Context, name reference.Named) (distribution.Repository, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r.registry.Repository(ctx, name)
}
func (r *registry) Repositories(ctx context.Context, repos []string, last string) (n int, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	n, err = r.enumerator.EnumerateRepositories(ctx, repos, last)
	if err == errNoSpaceInSlice {
		return n, errors.New("client requested zero entries")
	}
	return
}
func (r *registry) Blobs() distribution.BlobEnumerator {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r.registry.Blobs()
}
func (r *registry) BlobStatter() distribution.BlobStatter {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r.registry.BlobStatter()
}

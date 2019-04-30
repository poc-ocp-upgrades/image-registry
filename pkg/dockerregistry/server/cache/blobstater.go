package cache

import (
	"context"
	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
)

type BlobStatter struct {
	Svc	distribution.BlobStatter
	Cache	DigestCache
}

var _ distribution.BlobStatter = &BlobStatter{}

func (bs *BlobStatter) Stat(ctx context.Context, dgst digest.Digest) (distribution.Descriptor, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	desc, err := bs.Cache.Get(dgst)
	if err == nil || err != distribution.ErrBlobUnknown || bs.Svc == nil {
		return desc, err
	}
	desc, err = bs.Svc.Stat(ctx, dgst)
	if err != nil {
		return desc, err
	}
	_ = bs.Cache.Add(dgst, &DigestValue{desc: &desc})
	return desc, nil
}

package wrapped

import (
	"context"
	"github.com/docker/distribution"
)

type tagService struct {
	tagService	distribution.TagService
	wrapper		Wrapper
}

var _ distribution.TagService = &tagService{}

func NewTagService(ts distribution.TagService, wrapper Wrapper) distribution.TagService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &tagService{tagService: ts, wrapper: wrapper}
}
func (ts *tagService) Get(ctx context.Context, tag string) (desc distribution.Descriptor, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	err = ts.wrapper(ctx, "TagService.Get", func(ctx context.Context) error {
		desc, err = ts.tagService.Get(ctx, tag)
		return err
	})
	return
}
func (ts *tagService) Tag(ctx context.Context, tag string, desc distribution.Descriptor) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return ts.wrapper(ctx, "TagService.Tag", func(ctx context.Context) error {
		return ts.tagService.Tag(ctx, tag, desc)
	})
}
func (ts *tagService) Untag(ctx context.Context, tag string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return ts.wrapper(ctx, "TagService.Untag", func(ctx context.Context) error {
		return ts.tagService.Untag(ctx, tag)
	})
}
func (ts *tagService) All(ctx context.Context) (tags []string, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	err = ts.wrapper(ctx, "TagService.All", func(ctx context.Context) error {
		tags, err = ts.tagService.All(ctx)
		return err
	})
	return
}
func (ts *tagService) Lookup(ctx context.Context, desc distribution.Descriptor) (tags []string, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	err = ts.wrapper(ctx, "TagService.Lookup", func(ctx context.Context) error {
		tags, err = ts.tagService.Lookup(ctx, desc)
		return err
	})
	return
}

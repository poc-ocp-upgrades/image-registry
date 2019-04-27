package errors

import (
	"context"
	dcontext "github.com/docker/distribution/context"
	errcode "github.com/docker/distribution/registry/api/errcode"
)

type errMessageKey struct{}

func (errMessageKey) String() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return "err.message"
}

type errDetailKey struct{}

func (errDetailKey) String() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return "err.detail"
}
func Handle(ctx context.Context, message string, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	ctx = context.WithValue(ctx, errMessageKey{}, err)
	switch err := err.(type) {
	case errcode.Error:
		ctx = context.WithValue(ctx, errDetailKey{}, err.Detail)
	}
	dcontext.GetLogger(ctx, errMessageKey{}, errDetailKey{}).Error(message)
}

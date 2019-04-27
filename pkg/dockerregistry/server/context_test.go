package server

import "context"

func withAppMiddleware(parent context.Context, am appMiddleware) context.Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return context.WithValue(parent, appMiddlewareKey, am)
}

package server

import (
	"context"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/client"
)

type contextKey string

const (
	appMiddlewareKey	contextKey	= "appMiddleware"
	userClientKey		contextKey	= "userClient"
	authPerformedKey	contextKey	= "authPerformed"
	deferredErrorsKey	contextKey	= "deferredErrors"
)

func appMiddlewareFrom(ctx context.Context) appMiddleware {
	_logClusterCodePath()
	defer _logClusterCodePath()
	am, _ := ctx.Value(appMiddlewareKey).(appMiddleware)
	return am
}
func withUserClient(parent context.Context, userClient client.Interface) context.Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return context.WithValue(parent, userClientKey, userClient)
}
func userClientFrom(ctx context.Context) (client.Interface, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	userClient, ok := ctx.Value(userClientKey).(client.Interface)
	return userClient, ok
}
func withAuthPerformed(parent context.Context) context.Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return context.WithValue(parent, authPerformedKey, true)
}
func authPerformed(ctx context.Context) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	authPerformed, ok := ctx.Value(authPerformedKey).(bool)
	return ok && authPerformed
}
func withDeferredErrors(parent context.Context, errs deferredErrors) context.Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return context.WithValue(parent, deferredErrorsKey, errs)
}
func deferredErrorsFrom(ctx context.Context) (deferredErrors, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	errs, ok := ctx.Value(deferredErrorsKey).(deferredErrors)
	return errs, ok
}

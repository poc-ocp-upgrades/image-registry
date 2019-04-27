package metrics

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/docker/distribution/registry/handlers"
	gorillahandlers "github.com/gorilla/handlers"
)

func Dispatcher(ctx *handlers.Context, r *http.Request) http.Handler {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return gorillahandlers.MethodHandler{"GET": prometheus.Handler()}
}

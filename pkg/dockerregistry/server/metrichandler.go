package server

import (
	"net/http"
	"github.com/docker/distribution/registry/auth"
	"github.com/docker/distribution/registry/handlers"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/api"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/metrics"
)

func RegisterMetricHandler(app *handlers.App) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	getMetricsAccess := func(r *http.Request) []auth.Access {
		return []auth.Access{{Resource: auth.Resource{Type: "metrics"}, Action: "get"}}
	}
	extensionsRouter := app.NewRoute().PathPrefix(api.ExtensionsPrefix).Subrouter()
	app.RegisterRoute("extensions-metrics", extensionsRouter.Path(api.MetricsPath).Methods("GET"), metrics.Dispatcher, handlers.NameNotRequired, getMetricsAccess)
}

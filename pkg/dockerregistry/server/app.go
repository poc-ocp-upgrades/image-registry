package server

import (
	"context"
	"net/http"
	"time"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/docker/distribution"
	"github.com/docker/distribution/configuration"
	dcontext "github.com/docker/distribution/context"
	registrycache "github.com/docker/distribution/registry/storage/cache"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	kubecache "k8s.io/apimachinery/pkg/util/cache"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/cache"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/client"
	registryconfig "github.com/openshift/image-registry/pkg/dockerregistry/server/configuration"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/maxconnections"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/metrics"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/supermiddleware"
)

const (
	defaultDescriptorCacheSize		= 6 * 4096
	defaultDigestToRepositoryCacheSize	= 2048
	defaultPaginationCacheSize		= 1024
)

type appMiddleware interface {
	Apply(supermiddleware.App) supermiddleware.App
}
type App struct {
	ctx		context.Context
	registryClient	client.RegistryClient
	config		*registryconfig.Configuration
	writeLimiter	maxconnections.Limiter
	driver		storagedriver.StorageDriver
	registry	distribution.Namespace
	quotaEnforcing	*quotaEnforcingConfig
	cache		cache.DigestCache
	metrics		metrics.Metrics
	paginationCache	*kubecache.LRUExpireCache
}

func (app *App) Storage(driver storagedriver.StorageDriver, options map[string]interface{}) (storagedriver.StorageDriver, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	app.driver = app.metrics.StorageDriver(driver)
	return app.driver, nil
}
func (app *App) Registry(nm distribution.Namespace, options map[string]interface{}) (distribution.Namespace, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	app.registry = nm
	return &registry{registry: nm, enumerator: NewCachingRepositoryEnumerator(app.registryClient, app.paginationCache)}, nil
}
func (app *App) BlobStatter() distribution.BlobStatter {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &cache.BlobStatter{Cache: app.cache, Svc: app.registry.BlobStatter()}
}
func (app *App) CacheProvider(ctx context.Context, options map[string]interface{}) (registrycache.BlobDescriptorCacheProvider, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &cache.Provider{Cache: app.cache}, nil
}
func NewApp(ctx context.Context, registryClient client.RegistryClient, dockerConfig *configuration.Configuration, extraConfig *registryconfig.Configuration, writeLimiter maxconnections.Limiter) http.Handler {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	app := &App{ctx: ctx, registryClient: registryClient, config: extraConfig, writeLimiter: writeLimiter, quotaEnforcing: newQuotaEnforcingConfig(ctx, extraConfig.Quota), paginationCache: kubecache.NewLRUExpireCache(defaultPaginationCacheSize)}
	if app.config.Metrics.Enabled {
		app.metrics = metrics.NewMetrics(metrics.NewPrometheusSink())
	} else {
		app.metrics = metrics.NewNoopMetrics()
	}
	cacheTTL := time.Duration(0)
	if !app.config.Cache.Disabled {
		cacheTTL = app.config.Cache.BlobRepositoryTTL
	}
	digestCache, err := cache.NewBlobDigest(defaultDescriptorCacheSize, defaultDigestToRepositoryCacheSize, cacheTTL, app.metrics)
	if err != nil {
		dcontext.GetLogger(ctx).Fatalf("unable to create cache: %v", err)
	}
	app.cache = digestCache
	superapp := supermiddleware.App(app)
	if am := appMiddlewareFrom(ctx); am != nil {
		superapp = am.Apply(superapp)
	}
	dockerApp := supermiddleware.NewApp(ctx, dockerConfig, superapp)
	if app.driver == nil {
		dcontext.GetLogger(ctx).Fatalf("configuration error: the storage driver middleware %q is not activated", supermiddleware.Name)
	}
	if app.registry == nil {
		dcontext.GetLogger(ctx).Fatalf("configuration error: the registry middleware %q is not activated", supermiddleware.Name)
	}
	if dockerConfig.Auth.Type() == supermiddleware.Name {
		tokenRealm, err := registryconfig.TokenRealm(extraConfig.Auth.TokenRealm)
		if err != nil {
			dcontext.GetLogger(dockerApp).Fatalf("error setting up token auth: %s", err)
		}
		err = dockerApp.NewRoute().Methods("GET").PathPrefix(tokenRealm.Path).Handler(NewTokenHandler(ctx, registryClient)).GetError()
		if err != nil {
			dcontext.GetLogger(dockerApp).Fatalf("error setting up token endpoint at %q: %v", tokenRealm.Path, err)
		}
		dcontext.GetLogger(dockerApp).Debugf("configured token endpoint at %q", tokenRealm.String())
	}
	app.registerBlobHandler(dockerApp)
	isImageClient, err := registryClient.Client()
	if err != nil {
		dcontext.GetLogger(dockerApp).Fatalf("unable to get client for signatures: %v", err)
	}
	RegisterSignatureHandler(dockerApp, isImageClient)
	if dockerApp.Config.HTTP.Headers == nil {
		dockerApp.Config.HTTP.Headers = http.Header{}
	}
	dockerApp.Config.HTTP.Headers.Set("X-Registry-Supports-Signatures", "1")
	dockerApp.RegisterHealthChecks()
	h := http.Handler(dockerApp)
	if extraConfig.Metrics.Enabled {
		RegisterMetricHandler(dockerApp)
		h = promhttp.InstrumentHandlerCounter(metrics.HTTPRequestsTotal, h)
		h = promhttp.InstrumentHandlerDuration(metrics.HTTPRequestDurationSeconds, h)
		h = promhttp.InstrumentHandlerInFlight(metrics.HTTPInFlightRequests, h)
		h = promhttp.InstrumentHandlerRequestSize(metrics.HTTPRequestSizeBytes, h)
		h = promhttp.InstrumentHandlerResponseSize(metrics.HTTPResponseSizeBytes, h)
		h = promhttp.InstrumentHandlerTimeToWriteHeader(metrics.HTTPTimeToWriteHeaderSeconds, h)
	}
	dcontext.GetLogger(dockerApp).Infof("Using %q as Docker Registry URL", extraConfig.Server.Addr)
	return h
}

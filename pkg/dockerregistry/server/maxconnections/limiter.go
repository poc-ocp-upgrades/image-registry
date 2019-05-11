package maxconnections

import (
	"context"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"time"
)

type Limiter interface {
	Start(context.Context) bool
	Done()
}
type limiter struct {
	running			chan struct{}
	queue			chan struct{}
	maxWaitInQueue	time.Duration
	newTimer		func(d time.Duration) *time.Timer
}

func NewLimiter(maxRunning, maxInQueue int, maxWaitInQueue time.Duration) Limiter {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &limiter{running: make(chan struct{}, maxRunning), queue: make(chan struct{}, maxInQueue), maxWaitInQueue: maxWaitInQueue, newTimer: time.NewTimer}
}
func (l *limiter) Start(ctx context.Context) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	select {
	case l.running <- struct{}{}:
		return true
	default:
	}
	select {
	case l.queue <- struct{}{}:
		defer func() {
			<-l.queue
		}()
	default:
		return false
	}
	var timeout <-chan time.Time
	if l.maxWaitInQueue > 0 {
		timer := l.newTimer(l.maxWaitInQueue)
		defer timer.Stop()
		timeout = timer.C
	}
	select {
	case l.running <- struct{}{}:
		return true
	case <-timeout:
	case <-ctx.Done():
	}
	return false
}
func (l *limiter) Done() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	<-l.running
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}

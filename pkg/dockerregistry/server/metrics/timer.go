package metrics

import (
	"time"
)

type Timer interface{ Stop() }

func NewTimer(observer Observer) Timer {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &timer{observer: observer, startTime: time.Now()}
}

type timer struct {
	observer	Observer
	startTime	time.Time
}

func (t *timer) Stop() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	t.observer.Observe(time.Since(t.startTime).Seconds())
}

package maxconnections

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/openshift/image-registry/pkg/testutil/counter"
)

func TestMaxConnections(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	const timeout = 1 * time.Second
	maxRunning := 1
	maxInQueue := 2
	maxWaitInQueue := time.Duration(0)
	lim := NewLimiter(maxRunning, maxInQueue, maxWaitInQueue)
	handlerBarrier := make(chan struct{}, maxRunning+maxInQueue+1)
	h := New(lim, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-handlerBarrier
		http.Error(w, "OK", http.StatusOK)
	}))
	ts := httptest.NewServer(h)
	defer ts.Close()
	defer func() {
		close(handlerBarrier)
	}()
	c := counter.New()
	done := make(chan struct{})
	wait := func(reason string) {
		select {
		case <-done:
		case <-time.After(timeout):
			t.Fatal(reason)
		}
	}
	for i := 0; i < maxRunning+maxInQueue+1; i++ {
		go func() {
			res, err := http.Get(ts.URL)
			if err != nil {
				t.Errorf("failed to get %s: %s", ts.URL, err)
			}
			c.Add(res.StatusCode, 1)
			done <- struct{}{}
		}()
	}
	wait("timeout while waiting one failed client")
	if diff := c.Diff(counter.M{429: 1}); diff != nil {
		t.Error(diff)
	}
	handlerBarrier <- struct{}{}
	wait("timeout while waiting the first succeed client")
	if diff := c.Diff(counter.M{200: 1, 429: 1}); diff != nil {
		t.Error(diff)
	}
	handlerBarrier <- struct{}{}
	wait("timeout while waiting the second succeed client")
	if diff := c.Diff(counter.M{200: 2, 429: 1}); diff != nil {
		t.Error(diff)
	}
	handlerBarrier <- struct{}{}
	wait("timeout while waiting the third succeed client")
	if diff := c.Diff(counter.M{200: 3, 429: 1}); diff != nil {
		t.Error(diff)
	}
}

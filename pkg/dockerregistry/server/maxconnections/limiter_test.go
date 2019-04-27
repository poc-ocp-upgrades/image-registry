package maxconnections

import (
	"context"
	"testing"
	"time"
	"github.com/openshift/image-registry/pkg/testutil/counter"
)

func TestLimiter(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	const timeout = 1 * time.Second
	maxRunning := 2
	maxInQueue := 3
	maxWaitInQueue := time.Duration(1)
	lim := NewLimiter(maxRunning, maxInQueue, maxWaitInQueue)
	deadline := make(chan time.Time)
	lim.(*limiter).newTimer = func(d time.Duration) *time.Timer {
		t := time.NewTimer(d)
		t.C = deadline
		return t
	}
	ctx := context.Background()
	c := counter.New()
	jobBarrier := make(chan struct{}, maxRunning+maxInQueue+1)
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
			started := lim.Start(ctx)
			defer func() {
				c.Add(started, 1)
				done <- struct{}{}
			}()
			if started {
				<-jobBarrier
				lim.Done()
			}
		}()
	}
	wait("timeout while waiting one failed job")
	if diff := c.Diff(counter.M{false: 1}); diff != nil {
		t.Error(diff)
	}
	jobBarrier <- struct{}{}
	wait("timeout while waiting one succeed job")
	if diff := c.Diff(counter.M{false: 1, true: 1}); diff != nil {
		t.Error(diff)
	}
	close(deadline)
	wait("timeout while waiting the first failed job from the queue")
	wait("timeout while waiting the second failed job from the queue")
	if diff := c.Diff(counter.M{false: 3, true: 1}); diff != nil {
		t.Error(diff)
	}
	jobBarrier <- struct{}{}
	jobBarrier <- struct{}{}
	wait("timeout while waiting the first succeed job")
	wait("timeout while waiting the second succeed job")
	if diff := c.Diff(counter.M{false: 3, true: 3}); diff != nil {
		t.Error(diff)
	}
}
func TestLimiterContext(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	const timeout = 1 * time.Second
	maxRunning := 2
	maxInQueue := 3
	maxWaitInQueue := 120 * time.Second
	lim := NewLimiter(maxRunning, maxInQueue, maxWaitInQueue)
	type job struct {
		ctx		context.Context
		cancel		context.CancelFunc
		finished	bool
	}
	c := counter.New()
	jobs := make(chan *job, maxRunning+maxInQueue+1)
	jobBarrier := make(chan struct{}, maxRunning+maxInQueue+1)
	done := make(chan struct{})
	startJobs := func(amount int) {
		for i := 0; i < amount; i++ {
			go func() {
				ctx, cancel := context.WithCancel(context.Background())
				job := &job{ctx: ctx, cancel: cancel, finished: false}
				jobs <- job
				started := lim.Start(ctx)
				defer func() {
					c.Add(started, 1)
					job.finished = true
					done <- struct{}{}
				}()
				if started {
					<-jobBarrier
					lim.Done()
				}
			}()
		}
	}
	cancelJobs := func(amount int, desc string) {
		i := 0
		for i < amount {
			select {
			case job := <-jobs:
				if job.finished {
					continue
				}
				job.cancel()
				i++
			case <-time.After(timeout):
				t.Fatalf("timeout while cancelling %s (%d of %d)", desc, i+1, amount)
			}
		}
	}
	finishJobs := func(amount int, desc string) {
		for i := 0; i < amount; i++ {
			select {
			case jobBarrier <- struct{}{}:
			case <-time.After(timeout):
				t.Fatalf("timeout while finishing %s (%d of %d)", desc, i+1, amount)
			}
		}
	}
	waitJobs := func(amount int, desc string) {
		for i := 0; i < amount; i++ {
			select {
			case <-done:
			case <-time.After(timeout):
				t.Fatalf("timeout while waiting %s (%d of %d)", desc, i+1, amount)
			}
		}
	}
	startJobs(maxRunning + maxInQueue + 1)
	waitJobs(1, "the job that doesn't fit in the queue from the first portion")
	if diff := c.Diff(counter.M{false: 1}); diff != nil {
		t.Error(diff)
	}
	cancelJobs(maxRunning+maxInQueue, "the jobs from the first portion")
	waitJobs(maxInQueue, "the cancelled jobs from the queue from the first portion")
	if diff := c.Diff(counter.M{false: 4}); diff != nil {
		t.Error(diff)
	}
	startJobs(maxInQueue + 1)
	waitJobs(1, "the job that doesn't fit in the queue from the second portion")
	if diff := c.Diff(counter.M{false: 5}); diff != nil {
		t.Error(diff)
	}
	finishJobs(maxRunning+maxInQueue, "all running and queued jobs")
	waitJobs(maxRunning+maxInQueue, "all finished jobs")
	if diff := c.Diff(counter.M{false: 5, true: 5}); diff != nil {
		t.Error(diff)
	}
}

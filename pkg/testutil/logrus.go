package testutil

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"
	dcontext "github.com/docker/distribution/context"
	"github.com/sirupsen/logrus"
)

type logrusHook struct{ t *testing.T }

func (h *logrusHook) Levels() []logrus.Level {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return logrus.AllLevels
}
func (h *logrusHook) Fire(e *logrus.Entry) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	line, err := e.String()
	if err != nil {
		h.t.Logf("unable to read entry: %v", err)
		return err
	}
	line = strings.TrimRight(line, " \n")
	h.t.Log(line)
	return nil
}
func WithTestLogger(parent context.Context, t *testing.T) context.Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	log := logrus.New()
	log.Level = logrus.DebugLevel
	log.Out = ioutil.Discard
	log.Hooks.Add(&logrusHook{t: t})
	return dcontext.WithLogger(parent, logrus.NewEntry(log))
}

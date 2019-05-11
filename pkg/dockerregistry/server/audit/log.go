package audit

import (
	"context"
	"sync"
	dcontext "github.com/docker/distribution/context"
	"github.com/sirupsen/logrus"
)

const (
	LogEntryType		= "openshift.logger"
	AuditUserEntry		= "openshift.auth.user"
	AuditUserIDEntry	= "openshift.auth.userid"
	AuditStatusEntry	= "openshift.request.status"
	AuditErrorEntry		= "openshift.request.error"
	auditLoggerKey		= "openshift.audit.logger"
	DefaultLoggerType	= "registry"
	AuditLoggerType		= "audit"
	OpStatusBegin		= "begin"
	OpStatusError		= "error"
	OpStatusOK			= "success"
)

type Logger struct {
	mu		sync.Mutex
	ctx		context.Context
	logger	*logrus.Logger
}

func NewLogger(ctx context.Context) *Logger {
	_logClusterCodePath()
	defer _logClusterCodePath()
	logger := &Logger{logger: logrus.New(), ctx: ctx}
	if entry, ok := dcontext.GetLogger(ctx).(*logrus.Entry); ok {
		logger.SetFormatter(entry.Logger.Formatter)
	} else if lgr, ok := dcontext.GetLogger(ctx).(*logrus.Logger); ok {
		logger.SetFormatter(lgr.Formatter)
	}
	return logger
}
func (l *Logger) SetFormatter(formatter logrus.Formatter) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Formatter = formatter
}
func (l *Logger) Log(args ...interface{}) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	auditFields := logrus.Fields{LogEntryType: AuditLoggerType, AuditStatusEntry: OpStatusBegin}
	l.getEntry().WithFields(auditFields).Info(args...)
}
func (l *Logger) Logf(format string, args ...interface{}) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	auditFields := logrus.Fields{LogEntryType: AuditLoggerType}
	l.getEntry().WithFields(auditFields).Infof(format, args...)
}
func (l *Logger) LogResult(err error, args ...interface{}) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	auditFields := logrus.Fields{LogEntryType: AuditLoggerType, AuditStatusEntry: OpStatusOK}
	if err != nil {
		auditFields[AuditErrorEntry] = err
		auditFields[AuditStatusEntry] = OpStatusError
	}
	l.getEntry().WithFields(auditFields).Info(args...)
}
func (l *Logger) LogResultf(err error, format string, args ...interface{}) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	auditFields := logrus.Fields{LogEntryType: AuditLoggerType, AuditStatusEntry: OpStatusOK}
	if err != nil {
		auditFields[AuditErrorEntry] = err
		auditFields[AuditStatusEntry] = OpStatusError
	}
	l.getEntry().WithFields(auditFields).Infof(format, args...)
}
func (l *Logger) getEntry() *logrus.Entry {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if entry, ok := dcontext.GetLogger(l.ctx).(*logrus.Entry); ok {
		return l.logger.WithFields(entry.Data)
	}
	return logrus.NewEntry(l.logger)
}
func LoggerExists(ctx context.Context) (exists bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, exists = ctx.Value(auditLoggerKey).(*Logger)
	return
}
func GetLogger(ctx context.Context) *Logger {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if logger, ok := ctx.Value(auditLoggerKey).(*Logger); ok {
		return logger
	}
	return NewLogger(ctx)
}
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return context.WithValue(ctx, auditLoggerKey, logger)
}

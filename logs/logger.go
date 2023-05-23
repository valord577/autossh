package logs

import (
	"go.uber.org/zap"
)

type Logger struct {
	log   *zap.Logger
	suagr *zap.SugaredLogger
}

func (l *Logger) Sync() error {
	if l.log == nil {
		return nil
	}
	return l.log.Sync()
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	if l.log == nil {
		return
	}
	l.log.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	if l.log == nil {
		return
	}
	l.log.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	if l.log == nil {
		return
	}
	l.log.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	if l.log == nil {
		return
	}
	l.log.Error(msg, fields...)
}

func (l *Logger) DPanic(msg string, fields ...zap.Field) {
	if l.log == nil {
		return
	}
	l.log.DPanic(msg, fields...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	if l.suagr == nil {
		return
	}
	l.suagr.Debugf(template, args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	if l.suagr == nil {
		return
	}
	l.suagr.Infof(template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	if l.suagr == nil {
		return
	}
	l.suagr.Warnf(template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	if l.suagr == nil {
		return
	}
	l.suagr.Errorf(template, args...)
}

func (l *Logger) DPanicf(template string, args ...interface{}) {
	if l.suagr == nil {
		return
	}
	l.suagr.DPanicf(template, args...)
}

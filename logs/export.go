package logs

import (
	"os"
	"sync"

	"autossh/logs/internal"

	"go.uber.org/zap"
)

var (
	l    *Logger
	once sync.Once
)

func initZap() {
	once.Do(func() {
		log := internal.NewZapLogger(os.Stderr)
		l = &Logger{log: log, suagr: log.Sugar()}
	})
}

func With(fields ...zap.Field) *Logger {
	initZap()
	if l == nil {
		return nil
	}
	log := l.log.With(fields...)
	return &Logger{log: log, suagr: log.Sugar()}
}

// Sync calls *zap.Logger.Sync
func Sync() (err error) {
	if l == nil {
		return
	}
	return l.Sync()
}

// Debug calls *zap.Logger.Debug
func Debug(msg string, fields ...zap.Field) {
	initZap()
	if l == nil {
		return
	}
	l.Debug(msg, fields...)
}

// Info calls *zap.Logger.Info
func Info(msg string, fields ...zap.Field) {
	initZap()
	if l == nil {
		return
	}
	l.Info(msg, fields...)
}

// Warn calls *zap.Logger.Warn
func Warn(msg string, fields ...zap.Field) {
	initZap()
	if l == nil {
		return
	}
	l.Warn(msg, fields...)
}

// Error calls *zap.Logger.Error
func Error(msg string, fields ...zap.Field) {
	initZap()
	if l == nil {
		return
	}
	l.Error(msg, fields...)
}

// DPanic calls *zap.Logger.DPanic
func DPanic(msg string, fields ...zap.Field) {
	initZap()
	if l == nil {
		return
	}
	l.DPanic(msg, fields...)
}

// Debugf calls *zap.SugaredLogger.Debugf
func Debugf(template string, args ...interface{}) {
	initZap()
	if l == nil {
		return
	}
	l.Debugf(template, args...)
}

// Infof calls *zap.SugaredLogger.Infof
func Infof(template string, args ...interface{}) {
	initZap()
	if l == nil {
		return
	}
	l.Infof(template, args...)
}

// Warnf calls *zap.SugaredLogger.Warnf
func Warnf(template string, args ...interface{}) {
	initZap()
	if l == nil {
		return
	}
	l.Warnf(template, args...)
}

// Errorf calls *zap.SugaredLogger.Errorf
func Errorf(template string, args ...interface{}) {
	initZap()
	if l == nil {
		return
	}
	l.Errorf(template, args...)
}

// DPanicf calls *zap.SugaredLogger.DPanicf
func DPanicf(template string, args ...interface{}) {
	initZap()
	if l == nil {
		return
	}
	l.DPanicf(template, args...)
}

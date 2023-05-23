package internal

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(ws zapcore.WriteSyncer) *zap.Logger {
	newZapEncoder := newZapEncoderFunc()

	tz, layout := timeZone(), timeFormat()
	timeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(tz).Format(layout))
	}

	encoder := newZapEncoder(
		zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "lv",
			NameKey:       "log",
			CallerKey:     "cal",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime:    timeEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
		},
	)

	level := zapcore.InfoLevel
	zapOptions := []zap.Option{
		zap.AddCaller(), zap.AddCallerSkip(1),
	}
	if isDebug() {
		level = zapcore.DebugLevel
		zapOptions = append(zapOptions, zap.Development())
	}

	levelEnabler := zap.LevelEnablerFunc(
		func(l zapcore.Level) bool {
			return level <= l && l <= zapcore.DPanicLevel
		},
	)

	zapCore := zapcore.NewCore(encoder, ws, levelEnabler)
	return zap.New(zapCore, zapOptions...)
}

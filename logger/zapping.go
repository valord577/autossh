package logs

import (
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// development trace
	autosshLogsDebug      = "AUTOSSH_LOGS_DEBUG"
	autosshLogsTimeFormat = "AUTOSSH_LOGS_TIME_FORMAT"
)

func isDebug() bool {
	env := os.Getenv(autosshLogsDebug)
	debug, _ := strconv.ParseBool(env)
	return debug
}

func tmfmt() string {
	env := os.Getenv(autosshLogsTimeFormat)
	if len(env) < 1 {
		env = "2006-01-02 15:04:05.000 -07:00"
	}
	return env
}

var l *logger

func Init() {
	encoder := zapcore.NewConsoleEncoder(
		zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime:    zapcore.TimeEncoderOfLayout(tmfmt()),
			EncodeCaller:  zapcore.ShortCallerEncoder,
		},
	)

	level := zapcore.InfoLevel
	zapOptions := []zap.Option{
		zap.AddCaller(), zap.AddCallerSkip(2),
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
	zapCore := zapcore.NewCore(
		encoder, &logWriter{os.Stderr}, levelEnabler,
	)

	log := zap.New(zapCore, zapOptions...)
	l = &logger{log: log, suagr: log.Sugar()}
}

type logWriter struct {
	f *os.File
}

func (s *logWriter) Write(bs []byte) (n int, err error) {
	return s.f.Write(bs)
}

func (s *logWriter) Sync() error {
	return s.f.Sync()
}

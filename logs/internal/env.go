package internal

import (
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

// text (default) | json
const autosshLogsFormat = "AUTOSSH_LOGS_FORMAT"

func newZapEncoderFunc() func(zapcore.EncoderConfig) zapcore.Encoder {
	env := os.Getenv(autosshLogsFormat)
	env = strings.TrimSpace(env)
	switch env {
	case "json":
		return zapcore.NewJSONEncoder
	default:
		return zapcore.NewConsoleEncoder
	}
}

// development trace
const autosshLogsDebug = "AUTOSSH_LOGS_DEBUG"

func isDebug() bool {
	env := os.Getenv(autosshLogsDebug)
	debug, _ := strconv.ParseBool(env)
	return debug
}

// time zone
const autosshLogsTimeZone = "AUTOSSH_LOGS_TIME_ZONE"

func timeZone() *time.Location {
	env := os.Getenv(autosshLogsTimeZone)
	env = strings.TrimSpace(env)
	loc, err := time.LoadLocation(env)
	if err != nil {
		loc = time.Local
	}
	return loc
}

// Golang style time format template string.
// Default: "2006-01-02 15:04:05.000 -07:00"
const autosshLogsTimeFormat = "AUTOSSH_LOGS_TIME_FORMAT"

func timeFormat() string {
	env := os.Getenv(autosshLogsTimeFormat)
	env = strings.TrimSpace(env)
	if len(env) < 1 {
		env = "2006-01-02 15:04:05.000 -07:00"
	}
	return env
}

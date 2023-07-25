package logger

import (
	"github.com/rs/zerolog"
	"os"
)

const (
	LOG_LEVEL_DEBUG = "DEBUG"
	LOG_LEVEL_INFO  = "INFO"
	LOG_LEVEL_WARN  = "WARN"
	LOG_LEVEL_ERROR = "ERROR"
	LOG_LEVEL_FATAL = "FATAL"
	LOG_LEVEL_PANIC = "PANIC"
)

func SetupLogging() {
	zerolog.LevelFieldName = "level_name"
	zerolog.TimestampFieldName = "timestamp"
}

func NewLogger(component string) zerolog.Logger {

	level, ok := os.LookupEnv("MDL_COMN_LOGLEVEL")
	if !ok {
		level = LOG_LEVEL_INFO
	}

	levelValue := zerolog.InfoLevel

	switch level {
	case LOG_LEVEL_DEBUG:
		levelValue = zerolog.DebugLevel
	case LOG_LEVEL_WARN:
		levelValue = zerolog.WarnLevel
	case LOG_LEVEL_ERROR:
		levelValue = zerolog.ErrorLevel
	case LOG_LEVEL_FATAL:
		levelValue = zerolog.FatalLevel
	case LOG_LEVEL_PANIC:
		levelValue = zerolog.PanicLevel
	}
	logger := zerolog.New(os.Stderr).
		With().
		Str("component", component).
		Timestamp().
		Logger().
		Level(levelValue)

	return logger
}

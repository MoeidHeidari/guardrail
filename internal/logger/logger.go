package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init(verbose bool) {
	level := zapcore.InfoLevel
	if verbose {
		level = zapcore.DebugLevel
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      verbose,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	var err error
	log, err = cfg.Build()
	if err != nil {
		log = zap.NewNop()
	}
}

func L() *zap.Logger {
	if log == nil {
		return zap.NewNop()
	}
	return log
}

func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

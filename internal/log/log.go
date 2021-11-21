package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger initializes standard zap.Logger from the provided arguments.
// By default, level field is represented as numeric.
func NewLogger(level zapcore.Level, fields ...zapcore.Field) (*zap.Logger, error) {
	// Create a zap logger instance
	cfg := zapConfig(zap.NewAtomicLevelAt(level))

	return cfg.Build(zap.Fields(fields...))
}

func zapConfig(level zap.AtomicLevel) zap.Config {
	return zap.Config{
		Level:    level,
		Encoding: "json",
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			MessageKey:     "short_message",
			CallerKey:      "caller",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		DisableStacktrace: true,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
}

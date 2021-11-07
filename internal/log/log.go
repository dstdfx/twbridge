package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Opts represents options to init zap logger.
type Opts struct {
	// Level represents the minimum enabled logging level.
	Level zapcore.Level

	// Fields represents additional fields that will be applied.
	Fields []zapcore.Field
}

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
			StacktraceKey:  "full_message",
			CallerKey:      "_caller",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		DisableStacktrace: true,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
}

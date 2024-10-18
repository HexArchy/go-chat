package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(level string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	err := zapLevel.UnmarshalText([]byte(level))
	if err != nil {
		return nil, err
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	switch zapLevel {
	case zap.DebugLevel:
		config.Development = true
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.OutputPaths = []string{"stdout"}

	case zap.InfoLevel:
		config.DisableCaller = true
		config.DisableStacktrace = true

	case zap.WarnLevel:
		config.DisableCaller = false
		config.DisableStacktrace = true

	case zap.ErrorLevel:
		config.DisableCaller = false
		config.DisableStacktrace = false
	}

	return config.Build()
}

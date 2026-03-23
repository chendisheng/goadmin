package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level       string
	Format      string
	Output      string
	Development bool
}

func New(cfg Config) (*zap.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	zapCfg := zap.NewProductionConfig()
	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
	}

	if strings.EqualFold(strings.TrimSpace(cfg.Format), "console") {
		zapCfg.Encoding = "console"
	} else {
		zapCfg.Encoding = "json"
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)
	zapCfg.EncoderConfig.TimeKey = "ts"
	zapCfg.EncoderConfig.LevelKey = "level"
	zapCfg.EncoderConfig.NameKey = "logger"
	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.MessageKey = "msg"
	zapCfg.EncoderConfig.StacktraceKey = "stacktrace"
	zapCfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	output := strings.TrimSpace(cfg.Output)
	if output == "" {
		output = "stdout"
	}
	if output == "stdout" || output == "stderr" {
		zapCfg.OutputPaths = []string{output}
		zapCfg.ErrorOutputPaths = []string{"stderr"}
	} else {
		zapCfg.OutputPaths = []string{output}
		zapCfg.ErrorOutputPaths = []string{output}
	}

	options := []zap.Option{zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)}
	if cfg.Development {
		options = append(options, zap.Development())
	}

	logger, err := zapCfg.Build(options...)
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}
	return logger, nil
}

func parseLevel(level string) (zapcore.Level, error) {
	var parsed zapcore.Level
	if err := parsed.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(level)))); err != nil {
		return zapcore.InfoLevel, fmt.Errorf("parse logger level %q: %w", level, err)
	}
	return parsed, nil
}

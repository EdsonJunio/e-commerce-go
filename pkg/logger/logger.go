package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	baseLogger    *zap.Logger
	once          sync.Once
	defaultFields []zap.Field
)

// Init initializes the global zap logger with environment-based configuration.
// It should be called once at application startup.
func Init(cfg Config) error {
	var err error
	once.Do(func() {
		var zapCfg zap.Config

		if cfg.Environment == "production" {
			zapCfg = zap.NewProductionConfig()
		} else {
			zapCfg = zap.NewDevelopmentConfig()
			zapCfg.Encoding = "console"
			zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
			zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			zapCfg.EncoderConfig.EncodeName = zapcore.FullNameEncoder
		}

		zapCfg.EncoderConfig.TimeKey = "timestamp"

		baseLogger, err = zapCfg.Build(
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		if err != nil {
			return
		}

		// Redirect standard library log to zap
		_ = zap.RedirectStdLog(baseLogger)

		// Default structured fields (always attached to logs)
		defaultFields = []zap.Field{
			zap.String("service", cfg.Service),
			zap.String("env", cfg.Environment),
			zap.String("version", cfg.Version),
			zap.String("pid", fmt.Sprintf("%d", os.Getpid())),
		}

		baseLogger = baseLogger.With(defaultFields...)
	})
	return err
}

// Sync flushes any buffered log entries.
// It should be called with defer in main().
func Sync() error {
	if baseLogger != nil {
		return baseLogger.Sync()
	}
	return nil
}

// L returns the base logger instance.
func L() *zap.Logger {
	return baseLogger
}

// With returns a logger with additional contextual fields.
func With(fields ...zap.Field) *zap.Logger {
	return baseLogger.With(fields...)
}

package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	baseLogger *zap.Logger
	once       sync.Once
)

var defaultFields = []zap.Field{}

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

		baseLogger, err = zapCfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		if err != nil {
			return
		}

		zap.RedirectStdLog(baseLogger)

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

func Sync() error {
	if baseLogger != nil {
		return baseLogger.Sync()
	}
	return nil
}

func L() *zap.Logger {
	return baseLogger
}

func With(fields ...zap.Field) *zap.Logger {
	return baseLogger.With(fields...)
}

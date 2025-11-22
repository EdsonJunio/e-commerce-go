package transport

import (
	"e-commerce-go/pkg/logger"

	"go.uber.org/zap"
)

func LogByErrorMapping(mapping HTTPErrorMapping, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.String("error_code", mapping.Code))

	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	log := logger.L()

	switch mapping.LogLevel {
	case LevelError:
		log.Error(msg, fields...)
	case LevelWarn:
		log.Warn(msg, fields...)
	case LevelInfo:
		log.Info(msg, fields...)
	case LevelDebug:
		log.Debug(msg, fields...)
	default:
		log.Error(msg, fields...)
	}
}

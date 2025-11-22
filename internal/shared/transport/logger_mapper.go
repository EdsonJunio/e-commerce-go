package transport

import (
	"e-commerce-go/pkg/logger"

	"go.uber.org/zap"
)

func LogByErrorMapping(mapping HTTPErrorMapping, msg string, err error, fields ...zap.Field) {

	fields = append(fields, zap.String("error_code", mapping.Code))

	switch mapping.LogLevel {
	case "ERROR":
		logger.L().Error(msg, fields...)
	case "WARN":
		logger.L().Warn(msg, fields...)
	case "INFO":
		logger.L().Info(msg, fields...)
	default:
		logger.L().Debug(msg, fields...)
	}
}

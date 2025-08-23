package logger_test

import (
	"errors"
	"testing"

	"e-commerce-go/pkg/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestLoggerInitializationAndUsage(t *testing.T) {
	err := logger.Init(logger.Config{
		Environment: "development",
		Service:     "test-service",
		Version:     "v1.0.0",
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Simple log", func(t *testing.T) {
		logger.L().Info("Application started")
	})

	t.Run("Log with structured fields", func(t *testing.T) {
		logger.L().Info("User authenticated",
			zap.String("user_id", "123"),
			zap.String("email", "user@example.com"),
		)
	})

	t.Run("Error log with stack trace", func(t *testing.T) {
		err := simulateError()
		if err != nil {
			logger.L().Error("Failed to process request",
				zap.Error(err),
				zap.String("operation", "process_request"),
			)
		}
	})

	t.Run("Logger with fields using With", func(t *testing.T) {
		logger.With(
			zap.String("request_id", "abc123"),
			zap.String("service", "payment"),
		).Info("Payment processed")
	})
}

func simulateError() error {
	return errors.New("simulated error")
}

func TestLoggerWithZapTest(t *testing.T) {
	// Uses zaptest to create an isolated logger (no stdout/stderr pollution)
	zapLogger := zaptest.NewLogger(t)
	zapLogger.Info("Isolated test log")
	zapLogger.Warn("Test warning", zap.String("step", "setup"))
}

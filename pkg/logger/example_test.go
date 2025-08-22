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
		t.Fatalf("Erro ao inicializar o logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Log simples", func(t *testing.T) {
		logger.L().Info("Aplicação iniciada")
	})

	t.Run("Log com campos estruturados", func(t *testing.T) {
		logger.L().Info("Usuário autenticado",
			zap.String("user_id", "123"),
			zap.String("email", "usuario@exemplo.com"),
		)
	})

	t.Run("Log de erro com stack trace", func(t *testing.T) {
		err := simulateError()
		if err != nil {
			logger.L().Error("Falha ao processar requisição",
				zap.Error(err),
				zap.String("operation", "processar_requisicao"),
			)
		}
	})

	t.Run("Logger com campos via With", func(t *testing.T) {
		logger.With(
			zap.String("request_id", "abc123"),
			zap.String("service", "payment"),
		).Info("Pagamento processado")
	})
}

func simulateError() error {
	return errors.New("erro simulado")
}

func TestLoggerWithZapTest(t *testing.T) {
	// Usa zaptest para logger fake (isolado de stdout/stderr)
	zapLogger := zaptest.NewLogger(t)
	zapLogger.Info("Log de teste isolado")
	zapLogger.Warn("Aviso de teste", zap.String("etapa", "setup"))
}

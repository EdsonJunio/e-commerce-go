package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

func FromContext(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(ctxKey{}).(*zap.Logger)
	if !ok {
		return baseLogger
	}
	return l
}

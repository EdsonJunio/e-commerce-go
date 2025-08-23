package logger

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinLoggerMiddleware logs each HTTP request with latency and metadata.
// It uses the package-level baseLogger (initialized elsewhere).
func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", duration),
			zap.String("request_id", c.Writer.Header().Get("X-Request-ID")),
		}

		// Attach the last Gin error if present (useful for 4xx/5xx responses).
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.Last().Error()))
		}

		log := baseLogger.With(fields...)
		status := c.Writer.Status()

		switch {
		case status >= http.StatusInternalServerError:
			log.Error("server error")
		case status >= http.StatusBadRequest:
			log.Warn("client error")
		default:
			log.Info("request processed successfully")
		}
	}
}

// RecoveryWithLogger recovers from panics, logs the error and stack trace,
// and responds with 500 Internal Server Error in JSON.
func RecoveryWithLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				baseLogger.Error(
					"panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.Stack("stack"),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    "internal_error",
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

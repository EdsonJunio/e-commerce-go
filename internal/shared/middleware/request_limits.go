package middleware

import (
	"errors"
	"net/http"

	"e-commerce-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func LimitRequestBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

func RejectOversizedBody(c *gin.Context, err error) bool {
	var maxBytesError *http.MaxBytesError
	if !errors.As(err, &maxBytesError) {
		return false
	}
	response.Error(c, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the configured limit")
	return true
}

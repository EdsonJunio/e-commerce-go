package middleware

import (
	"net/http"
	"strings"

	"e-commerce-go/internal/shared/response"
	"e-commerce-go/internal/shared/service"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "user_id"
	ContextRole   = "role"
)

type AuthMiddleware struct {
	jwtService service.JWTService
}

func NewAuthMiddleware(jwtService service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.AbortWithError(c, http.StatusUnauthorized, "unauthenticated", "valid Bearer authentication is required")
			return
		}

		claims, err := m.jwtService.ValidateToken(parts[1])
		if err != nil {
			response.AbortWithError(c, http.StatusUnauthorized, "unauthenticated", "invalid or expired token")
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentRole, exists := c.Get(ContextRole)
		if !exists || currentRole != role {
			response.AbortWithError(c, http.StatusForbidden, "forbidden", "insufficient permissions")
			return
		}
		c.Next()
	}
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"e-commerce-go/internal/shared/response"
	"e-commerce-go/internal/shared/service"
)

type AuthMiddleware struct {
	jwtService service.JWTService
}

func NewAuthMiddleware(jwtService service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(
				c,
				http.StatusUnauthorized,
				"unauthorized",
				"Authorization header is missing",
			)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(
				c,
				http.StatusUnauthorized,
				"unauthorized",
				"Invalid authorization format. Use: Bearer <token>",
			)
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := m.jwtService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			response.Error(
				c,
				http.StatusUnauthorized,
				"unauthorized",
				"Invalid or expired token",
			)
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, ok := claims["user_id"].(float64); ok {
				c.Set("userID", int(userID))
			}
			if isAdmin, ok := claims["is_admin"].(bool); ok {
				c.Set("isAdmin", isAdmin)
			}
		}

		c.Next()
	}
}

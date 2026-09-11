package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"e-commerce-go/internal/shared/service"

	"github.com/gin-gonic/gin"
)

type jwtValidatorStub struct {
	claims *service.TokenClaims
	err    error
}

func (*jwtValidatorStub) GenerateToken(int, string) (string, error) { return "", nil }

func (s *jwtValidatorStub) ValidateToken(string) (*service.TokenClaims, error) {
	return s.claims, s.err
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		header     string
		validator  *jwtValidatorStub
		wantStatus int
	}{
		{name: "missing token", validator: &jwtValidatorStub{}, wantStatus: http.StatusUnauthorized},
		{name: "invalid token", header: "Bearer invalid", validator: &jwtValidatorStub{err: errors.New("invalid")}, wantStatus: http.StatusUnauthorized},
		{name: "valid token", header: "Bearer valid", validator: &jwtValidatorStub{claims: &service.TokenClaims{UserID: 9, Role: "admin"}}, wantStatus: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(NewAuthMiddleware(tt.validator).Handle())
			router.GET("/", func(c *gin.Context) {
				if c.GetInt(ContextUserID) != 9 || c.GetString(ContextRole) != "admin" {
					c.Status(http.StatusInternalServerError)
					return
				}
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.header)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", res.Code, tt.wantStatus, res.Body.String())
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{name: "admin", role: "admin", wantStatus: http.StatusNoContent},
		{name: "customer", role: "customer", wantStatus: http.StatusForbidden},
		{name: "missing", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.role != "" {
					c.Set(ContextRole, tt.role)
				}
			})
			router.Use(RequireRole("admin"))
			router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			res := httptest.NewRecorder()
			router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tt.wantStatus)
			}
		})
	}
}

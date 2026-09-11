package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"e-commerce-go/internal/identity/domain"

	"github.com/gin-gonic/gin"
)

type authServiceStub struct {
	token string
	err   error
}

func (s authServiceStub) Login(context.Context, string, string) (string, error) {
	return s.token, s.err
}

func TestAuthHandlerLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		service    authServiceStub
		body       string
		wantStatus int
		wantBody   string
	}{
		{name: "success", service: authServiceStub{token: "signed-token"}, body: `{"email":"user@example.com","password":"password"}`, wantStatus: http.StatusOK, wantBody: `"token":"signed-token"`},
		{name: "invalid credentials", service: authServiceStub{err: domain.ErrInvalidCredentials}, body: `{"email":"user@example.com","password":"wrong"}`, wantStatus: http.StatusUnauthorized, wantBody: `"code":"unauthorized"`},
		{name: "repository failure", service: authServiceStub{err: errors.New("database unavailable")}, body: `{"email":"user@example.com","password":"password"}`, wantStatus: http.StatusInternalServerError, wantBody: `"code":"internal_error"`},
		{name: "invalid payload", service: authServiceStub{}, body: `{"email":"invalid"}`, wantStatus: http.StatusBadRequest, wantBody: `"code":"invalid_request"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			NewAuthHandler(tt.service).RegisterRoutes(router)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", res.Code, tt.wantStatus, res.Body.String())
			}
			if !strings.Contains(res.Body.String(), tt.wantBody) {
				t.Fatalf("body = %s, want substring %s", res.Body.String(), tt.wantBody)
			}
		})
	}
}

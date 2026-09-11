package service

import (
	"strings"
	"testing"
	"time"

	"e-commerce-go/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTServiceRoundTrip(t *testing.T) {
	svc := newTestJWTService(t)
	token, err := svc.GenerateToken(42, "admin")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != 42 || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJWTServiceRejectsInvalidTokens(t *testing.T) {
	secret := strings.Repeat("s", 32)
	svc := newTestJWTService(t)
	now := time.Now().UTC()

	tests := []struct {
		name   string
		method jwt.SigningMethod
		claims TokenClaims
	}{
		{name: "expired", method: jwt.SigningMethodHS256, claims: claimsAt(now.Add(-time.Hour), now.Add(-time.Minute), "issuer", "audience")},
		{name: "wrong issuer", method: jwt.SigningMethodHS256, claims: claimsAt(now, now.Add(time.Hour), "other", "audience")},
		{name: "wrong audience", method: jwt.SigningMethodHS256, claims: claimsAt(now, now.Add(time.Hour), "issuer", "other")},
		{name: "wrong algorithm", method: jwt.SigningMethodHS384, claims: claimsAt(now, now.Add(time.Hour), "issuer", "audience")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwt.NewWithClaims(tt.method, tt.claims).SignedString([]byte(secret))
			if err != nil {
				t.Fatalf("sign token: %v", err)
			}
			if _, err := svc.ValidateToken(token); err == nil {
				t.Fatal("ValidateToken() expected an error")
			}
		})
	}
}

func TestJWTServiceRejectsWrongSignature(t *testing.T) {
	svc := newTestJWTService(t)
	now := time.Now().UTC()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsAt(now, now.Add(time.Hour), "issuer", "audience")).
		SignedString([]byte(strings.Repeat("x", 32)))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	if _, err := svc.ValidateToken(token); err == nil {
		t.Fatal("ValidateToken() expected an error")
	}
}

func TestNewJWTServiceRejectsUnsafeConfiguration(t *testing.T) {
	_, err := NewJWTService(config.JWTConfig{Secret: "short", Issuer: "issuer", Audience: "audience", AccessTokenTTL: time.Minute})
	if err == nil {
		t.Fatal("NewJWTService() expected an error")
	}
}

func newTestJWTService(t *testing.T) JWTService {
	t.Helper()
	svc, err := NewJWTService(config.JWTConfig{
		Secret: strings.Repeat("s", 32), Issuer: "issuer", Audience: "audience", AccessTokenTTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}
	return svc
}

func claimsAt(issuedAt, expiresAt time.Time, issuer, audience string) TokenClaims {
	return TokenClaims{
		UserID: 1,
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: issuer, Audience: jwt.ClaimStrings{audience},
			IssuedAt: jwt.NewNumericDate(issuedAt), NotBefore: jwt.NewNumericDate(issuedAt), ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
}

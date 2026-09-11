package service

import (
	"errors"
	"time"

	"e-commerce-go/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateToken(userID int, role string) (string, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
}

type jwtService struct {
	secretKey      []byte
	issuer         string
	audience       string
	accessTokenTTL time.Duration
	now            func() time.Time
}

func NewJWTService(cfg config.JWTConfig) (JWTService, error) {
	if len(cfg.Secret) < 32 {
		return nil, errors.New("JWT secret must contain at least 32 characters")
	}
	if cfg.Issuer == "" || cfg.Audience == "" || cfg.AccessTokenTTL <= 0 {
		return nil, errors.New("JWT issuer, audience, and access token TTL are required")
	}

	return &jwtService{
		secretKey:      []byte(cfg.Secret),
		issuer:         cfg.Issuer,
		audience:       cfg.Audience,
		accessTokenTTL: cfg.AccessTokenTTL,
		now:            time.Now,
	}, nil
}

func (s *jwtService) GenerateToken(userID int, role string) (string, error) {
	if userID <= 0 || role == "" {
		return "", errors.New("valid user ID and role are required")
	}

	now := s.now().UTC()
	claims := TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{s.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   "user",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *jwtService) ValidateToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected JWT signing method")
			}
			return s.secretKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.UserID <= 0 || claims.Role == "" {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

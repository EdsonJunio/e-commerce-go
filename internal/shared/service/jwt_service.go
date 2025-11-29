package service

import (
	"fmt"
	"time"

	"e-commerce-go/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	GenerateToken(userID int, isAdmin bool) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
}

type jwtService struct {
	secretKey string
	issuer    string
}

type jwtCustomClaim struct {
	UserID  int  `json:"user_id"`
	IsAdmin bool `json:"is_admin"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg *config.Config) JWTService {
	secret := "your-super-secret-key-that-nobody-knows"

	return &jwtService{
		secretKey: secret,
		issuer:    "ecommerce-api",
	}
}

func (s *jwtService) GenerateToken(userID int, isAdmin bool) (string, error) {

	claims := &jwtCustomClaim{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (s *jwtService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.secretKey), nil
	})
}

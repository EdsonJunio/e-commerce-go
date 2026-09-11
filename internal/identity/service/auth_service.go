package service

import (
	"context"
	"errors"

	"e-commerce-go/internal/identity/domain"
	"e-commerce-go/internal/shared/security"
	sharedservice "e-commerce-go/internal/shared/service"
)

type authService struct {
	repo       domain.UserRepository
	jwtService sharedservice.JWTService
}

func NewAuthService(repo domain.UserRepository, jwtService sharedservice.JWTService) domain.AuthService {
	return &authService{repo: repo, jwtService: jwtService}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}
	if !user.CanAuthenticate() {
		return "", domain.ErrInvalidCredentials
	}
	if err := security.CheckPasswordHash(password, user.PasswordHash); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return s.jwtService.GenerateToken(user.ID, string(user.Role))
}

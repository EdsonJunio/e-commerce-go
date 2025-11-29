package service

import (
	"e-commerce-go/internal/identity/domain"
	security "e-commerce-go/internal/shared/security"
	jwtService "e-commerce-go/internal/shared/service"
	"errors"
)

type authService struct {
	repo       domain.UserRepository
	jwtService jwtService.JWTService
}

func NewAuthService(repo domain.UserRepository, jwtSvc jwtService.JWTService) domain.AuthService {
	return &authService{
		repo:       repo,
		jwtService: jwtSvc,
	}
}

func (s *authService) Login(email, password string) (string, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := security.CheckPasswordHash(password, user.PasswordHash); err != nil {
		return "", errors.New("invalid credentials")
	}

	isAdmin := user.ID == 1
	token, err := s.jwtService.GenerateToken(user.ID, isAdmin)
	if err != nil {
		return "", err
	}

	return token, nil
}

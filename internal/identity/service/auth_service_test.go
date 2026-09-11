package service

import (
	"context"
	"errors"
	"testing"

	"e-commerce-go/internal/identity/domain"
	sharedservice "e-commerce-go/internal/shared/service"

	"golang.org/x/crypto/bcrypt"
)

type userRepositoryStub struct {
	user *domain.User
	err  error
}

func (s userRepositoryStub) GetByEmail(context.Context, string) (*domain.User, error) {
	return s.user, s.err
}

type jwtServiceStub struct {
	token  string
	userID int
	role   string
}

func (s *jwtServiceStub) GenerateToken(userID int, role string) (string, error) {
	s.userID = userID
	s.role = role
	return s.token, nil
}

func (*jwtServiceStub) ValidateToken(string) (*sharedservice.TokenClaims, error) {
	return nil, errors.New("not implemented")
}

func TestAuthServiceLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	tests := []struct {
		name      string
		repo      userRepositoryStub
		password  string
		wantToken string
		wantError error
	}{
		{name: "success", repo: userRepositoryStub{user: &domain.User{ID: 7, Role: domain.RoleAdmin, Status: domain.UserStatusActive, PasswordHash: string(hash)}}, password: "correct-password", wantToken: "token"},
		{name: "wrong password", repo: userRepositoryStub{user: &domain.User{ID: 7, Role: domain.RoleAdmin, Status: domain.UserStatusActive, PasswordHash: string(hash)}}, password: "wrong", wantError: domain.ErrInvalidCredentials},
		{name: "disabled user", repo: userRepositoryStub{user: &domain.User{ID: 7, Role: domain.RoleAdmin, Status: domain.UserStatusDisabled, PasswordHash: string(hash)}}, password: "correct-password", wantError: domain.ErrInvalidCredentials},
		{name: "unknown user", repo: userRepositoryStub{err: domain.ErrUserNotFound}, password: "correct-password", wantError: domain.ErrInvalidCredentials},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtStub := &jwtServiceStub{token: "token"}
			svc := NewAuthService(tt.repo, jwtStub)
			token, err := svc.Login(context.Background(), "user@example.com", tt.password)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("Login() error = %v, want %v", err, tt.wantError)
			}
			if token != tt.wantToken {
				t.Fatalf("Login() token = %q, want %q", token, tt.wantToken)
			}
			if tt.wantError == nil && (jwtStub.userID != 7 || jwtStub.role != "admin") {
				t.Fatalf("unexpected token arguments: userID=%d role=%q", jwtStub.userID, jwtStub.role)
			}
		})
	}
}

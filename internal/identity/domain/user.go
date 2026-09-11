package domain

import (
	"context"
	"time"
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type User struct {
	ID           int
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	Role         Role
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u User) CanAuthenticate() bool {
	return u.ID > 0 && u.Status == UserStatusActive
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
}

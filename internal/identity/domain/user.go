package domain

import "time"

type User struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRepository interface {
	GetByEmail(email string) (*User, error)
}

type AuthService interface {
	Login(email, password string) (string, error)
}

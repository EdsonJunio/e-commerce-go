package repository

import (
	"context"
	"errors"
	"time"

	"e-commerce-go/internal/identity/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

type userRecord struct {
	ID           int        `gorm:"column:id;primaryKey"`
	Email        string     `gorm:"column:email"`
	PasswordHash string     `gorm:"column:password_hash"`
	FullName     string     `gorm:"column:full_name"`
	Phone        string     `gorm:"column:phone"`
	Role         string     `gorm:"column:role"`
	Status       string     `gorm:"column:status"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (userRecord) TableName() string { return "users" }

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var record userRecord
	err := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?) AND deleted_at IS NULL", email).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:           record.ID,
		Email:        record.Email,
		PasswordHash: record.PasswordHash,
		FullName:     record.FullName,
		Phone:        record.Phone,
		Role:         domain.Role(record.Role),
		Status:       domain.UserStatus(record.Status),
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}, nil
}

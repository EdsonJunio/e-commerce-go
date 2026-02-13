package repository

import (
	"e-commerce-go/internal/catalog/domain"

	"gorm.io/gorm"
)

type SkuRepository struct {
	db *gorm.DB
}

func NewSkuRepository(db *gorm.DB) domain.SkuRepository {
	return &SkuRepository{db: db}
}

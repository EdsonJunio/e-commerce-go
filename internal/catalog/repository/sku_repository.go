package repository

import (
	"context"
	"e-commerce-go/internal/catalog/domain"

	"gorm.io/gorm"
)

type ProductSkuRepository struct {
	db *gorm.DB
}

func NewProductSkuRepository(db *gorm.DB) domain.ProductSkusRepository {
	return &ProductSkuRepository{db: db}
}

func (sr *ProductSkuRepository) List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]domain.Product_skus, int64, error) {
	var skus []domain.Product_skus
	var total int64

	tx := sr.db.WithContext(ctx).Model(&domain.Product_skus{})

	if skuID, ok := filters["sku_id"]; ok {
		tx = tx.Where("sku_id = ?", skuID)
	}
	if active, ok := filters["is_active"]; ok {
		tx = tx.Where("is_active = ?", active)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tx = tx.Order("id DESC")

	if err := tx.Offset(offset).Limit(limit).Find(&skus).Error; err != nil {
		return nil, 0, err
	}

	return skus, total, nil
}

package repository

import (
	"context"
	"errors"

	"e-commerce-go/internal/catalog/domain"

	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]domain.Product, int64, error) {
	var products []domain.Product
	var total int64

	tx := r.db.WithContext(ctx).Model(&domain.Product{})

	if catID, ok := filters["category_id"]; ok {
		tx = tx.Where("category_id = ?", catID)
	}
	if active, ok := filters["is_active"]; ok {
		tx = tx.Where("is_active = ?", active)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tx = tx.Order("id DESC")

	if err := tx.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) FindByID(ctx context.Context, id int) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).First(&product, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) FindBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&product).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	err := r.db.WithContext(ctx).Create(product).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrProductSlugExists
		}
		return err
	}

	return nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	err := r.db.WithContext(ctx).Save(product).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrProductSlugExists
		}
		return err
	}

	return nil
}

func (r *productRepository) Delete(ctx context.Context, id int) error {
	res := r.db.WithContext(ctx).Delete(&domain.Product{}, id)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

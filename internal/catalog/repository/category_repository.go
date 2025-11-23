package repository

import (
	"context"
	"errors"

	"e-commerce-go/internal/catalog/domain"

	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository returns a new category repository backed by GORM.
func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]domain.Category, int64, error) {
	var categories []domain.Category
	var total int64

	tx := r.db.WithContext(ctx).Model(&domain.Category{})

	if name, ok := filters["name"]; ok {
		tx = tx.Where("name = ?", name)
	}
	if active, ok := filters["is_active"]; ok {
		tx = tx.Where("is_active = ?", active)
	}
	if parentID, ok := filters["parent_id"]; ok {
		tx = tx.Where("parent_id = ?", parentID)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tx = tx.Order("id DESC")

	if err := tx.Offset(offset).Limit(limit).Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id int) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).First(&category, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrCategorySlugExists
		}
		return err
	}
	return nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	err := r.db.WithContext(ctx).Save(category).Error

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrCategorySlugExists
		}
		return err
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	res := r.db.WithContext(ctx).Delete(&domain.Category{}, id)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

package repository

import (
	"e-commerce-go/internal/catalog/domain"
	"errors"

	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository returns a new category repository backed by GORM.
// The caller is responsible for managing the lifecycle of db.
func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

// List returns categories with pagination and filters.
func (r *categoryRepository) List(limit, offset int, filters map[string]interface{}) ([]domain.Category, int64, error) {
	var categories []domain.Category
	var total int64

	tx := r.db.Model(&domain.Category{})
	for key, value := range filters {
		tx = tx.Where(key, value)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Offset(offset).Limit(limit).Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r categoryRepository) FindByID(id int) (*domain.Category, error) {
	var category domain.Category
	err := r.db.First(&category, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCategoryNotFound
	}

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r categoryRepository) FindBySlug(slug string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.Where("slug = ?", slug).First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r categoryRepository) Create(category *domain.Category) error {
	err := r.db.Create(category).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrCategorySlugExists
		}
		return err
	}
	return nil
}

func (r categoryRepository) Update(category *domain.Category) error {
	if category.ID == 0 {
		return gorm.ErrRecordNotFound
	}

	tx := r.db.Model(category).Updates(map[string]interface{}{
		"name":        category.Name,
		"slug":        category.Slug,
		"parent_id":   category.ParentID,
		"is_active":   category.IsActive,
		"description": category.Description,
	})

	if tx.Error != nil {
		if isUniqueViolation(tx.Error) {
			return domain.ErrCategorySlugRequired
		}
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r categoryRepository) Delete(id int) error {
	if id <= 0 {
		return gorm.ErrRecordNotFound
	}

	tx := r.db.Unscoped().Delete(&domain.Category{}, id)
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

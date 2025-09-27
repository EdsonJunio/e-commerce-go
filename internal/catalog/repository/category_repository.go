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
	//TODO implement me
	panic("implement me")
}

func (r categoryRepository) Update(category *domain.Category) error {
	//TODO implement me
	panic("implement me")
}

func (r categoryRepository) Delete(id int) error {
	//TODO implement me
	panic("implement me")
}

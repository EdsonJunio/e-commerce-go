package repository

import (
	"e-commerce-go/internal/catalog/domain"

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

func (c categoryRepository) List(limit, offset int, filters map[string]interface{}) ([]domain.Category, int64, error) {
	//TODO implement me
	panic("implement me")
}

func (c categoryRepository) FindByID(id int) (*domain.Category, error) {
	//TODO implement me
	panic("implement me")
}

func (c categoryRepository) FindBySlug(slug string) (*domain.Category, error) {
	//TODO implement me
	panic("implement me")
}

func (c categoryRepository) Create(category *domain.Category) error {
	//TODO implement me
	panic("implement me")
}

func (c categoryRepository) Update(category *domain.Category) error {
	//TODO implement me
	panic("implement me")
}

func (c categoryRepository) Delete(id int) error {
	//TODO implement me
	panic("implement me")
}

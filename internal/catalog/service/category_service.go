package service

import "e-commerce-go/internal/catalog/domain"

type categoryService struct {
	repo domain.CategoryRepository
}

func (c categoryService) List(limit, offset int, filters map[string]interface{}) ([]domain.Category, int64, error) {
	//TODO implement me
	panic("implement me")
}

func (c categoryService) FindByID(id int) (*domain.Category, error) {
	//TODO implement me
	panic("implement me")
}

func (c categoryService) FindBySlug(slug string) (*domain.Category, error) {
	//TODO implement me
	panic("implement me")
}

func (c categoryService) Create(category *domain.Category) error {
	//TODO implement me
	panic("implement me")
}

func (c categoryService) Update(category *domain.Category) error {
	//TODO implement me
	panic("implement me")
}

func (c categoryService) Delete(id int) error {
	//TODO implement me
	panic("implement me")
}

// NewCategoryService returns a new instance of categoryService implementing domain.CategoryRepository.
func NewCategoryService(repo domain.CategoryRepository) domain.CategoryRepository {
	return &categoryService{repo: repo}
}

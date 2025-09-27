package service

import (
	"e-commerce-go/internal/catalog/domain"
)

type categoryService struct {
	repo domain.CategoryRepository
}

// NewCategoryService returns a new instance of categoryService implementing domain.CategoryRepository.
func NewCategoryService(repo domain.CategoryRepository) domain.CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) ListCategories(p domain.Pagination, filters map[string]interface{}) ([]domain.Category, int64, error) {
	return s.repo.List(p.Limit, p.Offset, filters)
}

func (s categoryService) GetCategoryByID(id int) (*domain.Category, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidCategoryID
	}

	category, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if category == nil {
		return nil, domain.ErrCategoryNotFound
	}

	return category, nil
}

func (s categoryService) GetCategoryBySlug(slug string) (*domain.Category, error) {
	if slug == "" {
		return nil, domain.ErrCategoryDescriptionRequired
	}

	categorySlug, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	if categorySlug == nil {
		return nil, domain.ErrCategoryNotFound
	}

	return categorySlug, nil
}

func (s categoryService) CreateCategory(category *domain.Category) error {
	if category.Name == "" {
		return domain.ErrCategoryNameRequired
	}
	if category.Slug == "" {
		return domain.ErrCategorySlugRequired
	}
	if category.Description == "" {
		return domain.ErrCategoryDescriptionRequired
	}

	return s.repo.Create(category)

}

func (s categoryService) UpdateCategory(id int, category *domain.Category) error {
	//TODO implement me
	panic("implement me")
}

func (s categoryService) DeleteCategory(id int) error {
	//TODO implement me
	panic("implement me")
}

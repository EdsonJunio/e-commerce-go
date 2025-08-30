package service

import (
	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/pkg/logger"

	"go.uber.org/zap"
)

type categoryService struct {
	repo domain.CategoryRepository
}

// NewCategoryService returns a new instance of categoryService implementing domain.CategoryRepository.
func NewCategoryService(repo domain.CategoryRepository) domain.CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) ListCategories(p domain.Pagination, filters map[string]interface{}) ([]domain.Category, int64, error) {
	logger.L().Info(
		"listing categories",
		zap.Int("limit", p.Limit),
		zap.Int("page", p.Page),
		zap.Int("offset", p.Offset),
		zap.Any("filters", filters),
	)

	return s.repo.List(p.Limit, p.Offset, filters)
}

func (s categoryService) GetCategoryByID(id int) (*domain.Category, error) {
	//TODO implement me
	panic("implement me")
}

func (c categoryService) GetCategoryBySlug(slug string) (*domain.Category, error) {
	//TODO implement me
	panic("implement me")
}

func (s categoryService) CreateCategory(category *domain.Category) error {
	//TODO implement me
	panic("implement me")
}

func (s categoryService) UpdateCategory(id int, category *domain.Category) error {
	//TODO implement me
	panic("implement me")
}

func (s categoryService) DeleteCategory(id int) error {
	//TODO implement me
	panic("implement me")
}

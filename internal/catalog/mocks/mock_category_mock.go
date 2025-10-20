package mocks

import (
	"e-commerce-go/internal/catalog/domain"
	"errors"
)

type MockCategoryService struct {
	Categories []domain.Category
	Total      int64
	Err        error
}

func (m *MockCategoryService) ListCategories(p domain.Pagination, filters map[string]interface{}) ([]domain.Category, int64, error) {
	return m.Categories, m.Total, m.Err
}

func (m *MockCategoryService) GetCategoryByID(id int) (*domain.Category, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	for _, c := range m.Categories {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockCategoryService) GetCategoryBySlug(slug string) (*domain.Category, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	for _, c := range m.Categories {
		if c.Slug == slug {
			return &c, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockCategoryService) CreateCategory(category *domain.Category) error {
	return m.Err
}

func (m *MockCategoryService) UpdateCategory(id int, category *domain.Category) error {
	return m.Err
}

func (m *MockCategoryService) DeleteCategory(id int) error {
	return m.Err
}

package service

import (
	"e-commerce-go/internal/catalog/domain"
	"errors"
	"gorm.io/gorm"
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

	existing, err := s.repo.FindBySlug(category.Slug)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		return domain.ErrCategorySlugExists
	}

	if category.ParentID != nil && *category.ParentID > 0 {
		parent, err := s.repo.FindByID(*category.ParentID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if parent == nil {
			return domain.ErrParentCategoryNotFound
		}
	} else {
		var zeroID int
		category.ParentID = &zeroID
	}

	return s.repo.Create(category)
}

func (s categoryService) UpdateCategory(id int, category *domain.Category) error {
	if id <= 0 {
		return domain.ErrInvalidCategoryID
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if existing == nil {
		return domain.ErrCategoryNotFound
	}

	if category.Name != "" {
		existing.Name = category.Name
	}
	if category.Slug != "" {
		existing.Slug = category.Slug
	}
	if category.ParentID != nil {
		existing.ParentID = category.ParentID
	}
	if category.Description != "" {
		existing.Description = category.Description
	}

	if category.IsActive != existing.IsActive {
		existing.IsActive = category.IsActive
	}

	return s.repo.Update(existing)
}

func (s categoryService) DeleteCategory(id int) error {
	if id <= 0 {
		return domain.ErrInvalidCategoryID
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if existing == nil {
		return domain.ErrCategoryNotFound
	}

	return s.repo.Delete(id)
}

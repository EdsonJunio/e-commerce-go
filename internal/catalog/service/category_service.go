package service

import (
	"context"
	"e-commerce-go/internal/catalog/domain"
	"errors"
)

type categoryService struct {
	repo domain.CategoryRepository
}

// NewCategoryService returns a new instance of categoryService.
func NewCategoryService(
	repo domain.CategoryRepository) domain.CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) ListCategories(ctx context.Context, p domain.Pagination, filters domain.CategoryListFilters) ([]domain.Category, int64, error) {
	return s.repo.List(ctx, p.Limit, p.Offset, filters)
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	return s.findCategoryOrFail(ctx, id)
}

func (s *categoryService) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	if slug == "" {
		return nil, domain.ErrCategorySlugRequired
	}

	category, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, domain.ErrCategoryNotFound
	}

	return category, nil
}

func (s *categoryService) CreateCategory(ctx context.Context, category *domain.Category) error {
	if err := category.Validate(); err != nil {
		return err
	}

	if category.ParentID != nil && *category.ParentID > 0 {
		if err := s.ensureParentExists(ctx, *category.ParentID); err != nil {
			if errors.Is(err, domain.ErrCategoryNotFound) {
				return domain.ErrParentCategoryNotFound
			}
			return err
		}
	} else {
		category.ParentID = nil
	}

	return s.repo.Create(ctx, category)
}

func (s *categoryService) UpdateCategory(ctx context.Context, id int, changes domain.CategoryChanges) error {
	existing, err := s.findCategoryOrFail(ctx, id)
	if err != nil {
		return err
	}

	existing.UpdateState(changes)

	if err := existing.Validate(); err != nil {
		return err
	}

	if changes.ParentID != nil {
		if *changes.ParentID <= 0 || *changes.ParentID == id {
			return domain.ErrInvalidCategoryReference
		}
		if err := s.ensureParentDoesNotCycle(ctx, id, *changes.ParentID); err != nil {
			return err
		}
	}

	return s.repo.Update(ctx, existing)
}

func (s *categoryService) ensureParentDoesNotCycle(ctx context.Context, categoryID, parentID int) error {
	visited := make(map[int]struct{})
	for parentID != 0 {
		if parentID == categoryID {
			return domain.ErrInvalidCategoryReference
		}
		if _, seen := visited[parentID]; seen {
			return domain.ErrInvalidCategoryReference
		}
		visited[parentID] = struct{}{}

		nextParentID, err := s.repo.FindParentByID(ctx, parentID)
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return domain.ErrParentCategoryNotFound
		}
		if err != nil {
			return err
		}
		if nextParentID == nil {
			return nil
		}
		parentID = *nextParentID
	}
	return domain.ErrInvalidCategoryReference
}

func (s *categoryService) DeleteCategory(ctx context.Context, id int) error {
	if _, err := s.findCategoryOrFail(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *categoryService) findCategoryOrFail(ctx context.Context, id int) (*domain.Category, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidCategoryID
	}

	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, domain.ErrCategoryNotFound
	}

	return category, nil
}

func (s *categoryService) ensureParentExists(ctx context.Context, parentID int) error {
	parent, err := s.repo.FindByID(ctx, parentID)
	if err != nil {
		return err
	}
	if parent == nil {
		return domain.ErrParentCategoryNotFound
	}
	return nil
}

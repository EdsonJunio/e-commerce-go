package service

import (
	"context"
	"errors"
	"testing"

	"e-commerce-go/internal/catalog/domain"
)

type updateCategoryRepository struct {
	category  domain.Category
	updated   bool
	updateErr error
	parent    *domain.Category
}

func (*updateCategoryRepository) List(context.Context, int, int, domain.CategoryListFilters) ([]domain.Category, int64, error) {
	panic("not used")
}
func (r *updateCategoryRepository) FindByID(_ context.Context, id int) (*domain.Category, error) {
	if id != r.category.ID {
		if r.parent != nil && id == r.parent.ID {
			copy := *r.parent
			return &copy, nil
		}
		return nil, nil
	}
	copy := r.category
	return &copy, nil
}
func (*updateCategoryRepository) FindBySlug(context.Context, string) (*domain.Category, error) {
	panic("not used")
}
func (*updateCategoryRepository) Create(context.Context, *domain.Category) error { panic("not used") }
func (r *updateCategoryRepository) Update(_ context.Context, category *domain.Category) error {
	r.updated = true
	if r.updateErr != nil {
		return r.updateErr
	}
	r.category = *category
	return nil
}

func TestUpdateCategoryParentAndActiveMatrix(t *testing.T) {
	parentID := 2
	newParentID := 3
	falseValue := false
	trueValue := true
	tests := []struct {
		name       string
		changes    domain.CategoryChanges
		wantParent *int
		wantActive bool
	}{
		{"omitted", domain.CategoryChanges{}, &parentID, true},
		{"clear parent", domain.CategoryChanges{ClearParent: true}, nil, true},
		{"deactivate", domain.CategoryChanges{IsActive: &falseValue}, &parentID, false},
		{"assign different parent and activate", domain.CategoryChanges{ParentID: &newParentID, IsActive: &trueValue}, &newParentID, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &updateCategoryRepository{category: domain.Category{ID: 1, Name: "Name", Slug: "slug", Description: "Description", ParentID: &parentID, IsActive: true}, parent: &domain.Category{ID: 3}}
			if err := NewCategoryService(repo).UpdateCategory(context.Background(), 1, tt.changes); err != nil {
				t.Fatalf("update: %v", err)
			}
			if !repo.updated || repo.category.IsActive != tt.wantActive {
				t.Fatalf("category = %+v", repo.category)
			}
			if (repo.category.ParentID == nil) != (tt.wantParent == nil) {
				t.Fatalf("parent = %v, want %v", repo.category.ParentID, tt.wantParent)
			}
			if tt.wantParent != nil && *repo.category.ParentID != *tt.wantParent {
				t.Fatalf("parent = %d", *repo.category.ParentID)
			}
		})
	}
}

func TestUpdateCategoryRejectsInvalidChangesWithoutWrite(t *testing.T) {
	empty := " "
	missingParent := 3
	for _, tt := range []struct {
		name    string
		changes domain.CategoryChanges
		want    error
	}{
		{"empty name", domain.CategoryChanges{Name: &empty}, domain.ErrCategoryNameRequired},
		{"missing parent", domain.CategoryChanges{ParentID: &missingParent}, domain.ErrParentCategoryNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &updateCategoryRepository{category: domain.Category{ID: 1, Name: "Name", Slug: "slug", Description: "Description", IsActive: true}}
			err := NewCategoryService(repo).UpdateCategory(context.Background(), 1, tt.changes)
			if !errors.Is(err, tt.want) || repo.updated {
				t.Fatalf("error = %v, updated = %v", err, repo.updated)
			}
		})
	}
}

func TestUpdateCategoryPropagatesRepositoryFailure(t *testing.T) {
	want := errors.New("write failed")
	repo := &updateCategoryRepository{category: domain.Category{ID: 1, Name: "Name", Slug: "slug", Description: "Description"}, updateErr: want}
	if err := NewCategoryService(repo).UpdateCategory(context.Background(), 1, domain.CategoryChanges{}); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}
func (*updateCategoryRepository) Delete(context.Context, int) error { panic("not used") }

func TestUpdateCategoryPreservesOmittedActive(t *testing.T) {
	repo := &updateCategoryRepository{category: domain.Category{ID: 1, Name: "Before", Slug: "before", Description: "Description", IsActive: true}}
	name := "After"
	err := NewCategoryService(repo).UpdateCategory(context.Background(), 1, domain.CategoryChanges{Name: &name})
	if err != nil {
		t.Fatalf("update category: %v", err)
	}
	if !repo.updated || !repo.category.IsActive {
		t.Fatalf("omitted is_active deactivated category: %+v", repo.category)
	}
}

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
	ancestors map[int]domain.Category
	lookupErr error
}

func (*updateCategoryRepository) List(context.Context, int, int, domain.CategoryListFilters) ([]domain.Category, int64, error) {
	panic("not used")
}
func (r *updateCategoryRepository) FindByID(_ context.Context, id int) (*domain.Category, error) {
	if r.lookupErr != nil && id != r.category.ID {
		return nil, r.lookupErr
	}
	if id != r.category.ID {
		if ancestor, ok := r.ancestors[id]; ok {
			copy := ancestor
			return &copy, nil
		}
		if r.parent != nil && id == r.parent.ID {
			copy := *r.parent
			return &copy, nil
		}
		return nil, nil
	}
	copy := r.category
	return &copy, nil
}
func (r *updateCategoryRepository) FindParentByID(ctx context.Context, id int) (*int, error) {
	category, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, domain.ErrParentCategoryNotFound
	}
	return category.ParentID, nil
}

func TestUpdateCategoryRejectsIndirectCycle(t *testing.T) {
	rootID, childID, grandchildID := 1, 2, 3
	repo := &updateCategoryRepository{
		category: domain.Category{ID: rootID, Name: "Root", Slug: "root", Description: "Root"},
		ancestors: map[int]domain.Category{
			childID:      {ID: childID, ParentID: &rootID},
			grandchildID: {ID: grandchildID, ParentID: &childID},
		},
	}
	err := NewCategoryService(repo).UpdateCategory(context.Background(), rootID, domain.CategoryChanges{ParentID: &grandchildID})
	if !errors.Is(err, domain.ErrInvalidCategoryReference) {
		t.Fatalf("cycle error = %v, want invalid category reference", err)
	}
	if repo.updated {
		t.Fatal("cycle was persisted")
	}
}

func TestUpdateCategoryAncestorChain(t *testing.T) {
	rootID, childID, grandchildID, otherID := 1, 2, 3, 4
	readFailure := errors.New("ancestor read failed")
	tests := []struct {
		name      string
		parentID  int
		ancestors map[int]domain.Category
		lookupErr error
		wantErr   error
		wantWrite bool
	}{
		{"direct self-parent", rootID, nil, nil, domain.ErrInvalidCategoryReference, false},
		{"three-level cycle", grandchildID, map[int]domain.Category{grandchildID: {ID: grandchildID, ParentID: &childID}, childID: {ID: childID, ParentID: &rootID}}, nil, domain.ErrInvalidCategoryReference, false},
		{"valid chain", grandchildID, map[int]domain.Category{grandchildID: {ID: grandchildID, ParentID: &childID}, childID: {ID: childID, ParentID: &otherID}, otherID: {ID: otherID}}, nil, nil, true},
		{"missing ancestor", grandchildID, map[int]domain.Category{grandchildID: {ID: grandchildID, ParentID: &childID}}, nil, domain.ErrParentCategoryNotFound, false},
		{"lookup failure", grandchildID, nil, readFailure, readFailure, false},
		{"preexisting parent cycle", grandchildID, map[int]domain.Category{grandchildID: {ID: grandchildID, ParentID: &childID}, childID: {ID: childID, ParentID: &grandchildID}}, nil, domain.ErrInvalidCategoryReference, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &updateCategoryRepository{category: domain.Category{ID: rootID, Name: "Root", Slug: "root", Description: "Root"}, ancestors: tt.ancestors, lookupErr: tt.lookupErr}
			err := NewCategoryService(repo).UpdateCategory(context.Background(), rootID, domain.CategoryChanges{ParentID: &tt.parentID})
			if !errors.Is(err, tt.wantErr) || repo.updated != tt.wantWrite {
				t.Fatalf("error = %v, updated = %t; want %v, %t", err, repo.updated, tt.wantErr, tt.wantWrite)
			}
		})
	}
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

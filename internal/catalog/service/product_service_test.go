package service

import (
	"context"
	"errors"
	"testing"

	"e-commerce-go/internal/catalog/domain"
)

type updateProductRepository struct {
	product   domain.Product
	updated   bool
	findErr   error
	updateErr error
}

func (*updateProductRepository) List(context.Context, int, int, domain.ProductListFilters) ([]domain.Product, int64, error) {
	panic("not used")
}
func (r *updateProductRepository) FindByID(_ context.Context, id int) (*domain.Product, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if id != r.product.ID {
		return nil, nil
	}
	copy := r.product
	return &copy, nil
}
func (*updateProductRepository) FindBySlug(context.Context, string) (*domain.Product, error) {
	panic("not used")
}
func (*updateProductRepository) Create(context.Context, *domain.Product) error { panic("not used") }
func (r *updateProductRepository) Update(_ context.Context, product *domain.Product) error {
	r.updated = true
	if r.updateErr != nil {
		return r.updateErr
	}
	r.product = *product
	return nil
}

func TestUpdateProductActiveAndCategoryMatrix(t *testing.T) {
	oldCategory, newCategory := 2, 3
	falseValue, trueValue := false, true
	for _, tt := range []struct {
		name         string
		initial      bool
		changes      domain.ProductChanges
		wantActive   bool
		wantCategory int
	}{
		{"omitted", true, domain.ProductChanges{}, true, oldCategory},
		{"deactivate", true, domain.ProductChanges{IsActive: &falseValue}, false, oldCategory},
		{"activate and change category", false, domain.ProductChanges{IsActive: &trueValue, CategoryID: &newCategory}, true, newCategory},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := validProductRepository(oldCategory, tt.initial)
			categoryRepo := &updateCategoryRepository{category: domain.Category{ID: newCategory}}
			if err := NewProductService(repo, categoryRepo).UpdateProduct(context.Background(), 1, tt.changes); err != nil {
				t.Fatalf("update: %v", err)
			}
			if !repo.updated || repo.product.IsActive != tt.wantActive || *repo.product.CategoryID != tt.wantCategory {
				t.Fatalf("product = %+v", repo.product)
			}
		})
	}
}

func TestUpdateProductRejectsInvalidChangesWithoutWrite(t *testing.T) {
	categoryID := 2
	empty := " "
	zero := 0
	missing := 99
	for _, tt := range []struct {
		name    string
		changes domain.ProductChanges
		want    error
	}{
		{"empty name", domain.ProductChanges{Name: &empty}, domain.ErrProductNameRequired},
		{"invalid category", domain.ProductChanges{CategoryID: &zero}, domain.ErrInvalidCategoryReference},
		{"missing category", domain.ProductChanges{CategoryID: &missing}, domain.ErrInvalidCategoryReference},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := validProductRepository(categoryID, true)
			err := NewProductService(repo, &updateCategoryRepository{}).UpdateProduct(context.Background(), 1, tt.changes)
			if !errors.Is(err, tt.want) || repo.updated {
				t.Fatalf("error = %v, updated = %v", err, repo.updated)
			}
		})
	}
}

func TestUpdateProductPropagatesLookupAndWriteErrors(t *testing.T) {
	want := errors.New("repository failure")
	for _, tt := range []struct {
		name       string
		findErr    error
		updateErr  error
		wantUpdate bool
	}{
		{"lookup", want, nil, false},
		{"write", nil, want, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := validProductRepository(2, true)
			repo.findErr, repo.updateErr = tt.findErr, tt.updateErr
			err := NewProductService(repo, &updateCategoryRepository{}).UpdateProduct(context.Background(), 1, domain.ProductChanges{})
			if !errors.Is(err, want) || repo.updated != tt.wantUpdate || !repo.product.IsActive {
				t.Fatalf("error = %v, updated = %v, product = %+v", err, repo.updated, repo.product)
			}
		})
	}
}

func validProductRepository(categoryID int, active bool) *updateProductRepository {
	return &updateProductRepository{product: domain.Product{
		ID: 1, Name: "Before", Slug: "before", Description: "Description", SeoTitle: "Title",
		SeoDescription: "SEO description", CategoryID: &categoryID, IsActive: active,
	}}
}
func (*updateProductRepository) Delete(context.Context, int) error { panic("not used") }

func TestUpdateProductPreservesOmittedActive(t *testing.T) {
	categoryID := 2
	repo := &updateProductRepository{product: domain.Product{
		ID: 1, Name: "Before", Slug: "before", Description: "Description",
		SeoTitle: "Title", SeoDescription: "SEO description", CategoryID: &categoryID, IsActive: true,
	}}
	service := NewProductService(repo, &updateCategoryRepository{})
	name := "After"
	if err := service.UpdateProduct(context.Background(), 1, domain.ProductChanges{Name: &name}); err != nil {
		t.Fatalf("update product: %v", err)
	}
	if !repo.updated || !repo.product.IsActive {
		t.Fatalf("omitted is_active deactivated product: %+v", repo.product)
	}
}

package domain

import "testing"

func TestProductUpdateStatePreservesOmittedFieldsAndAppliesPresentFields(t *testing.T) {
	oldCategory := 2
	newCategory := 3
	product := Product{
		Name: "Name", Slug: "slug", Description: "Description", SeoTitle: "Title",
		SeoDescription: "SEO", CategoryID: &oldCategory, IsActive: true,
	}
	name := " New name "
	falseValue := false
	product.UpdateState(ProductChanges{Name: &name, CategoryID: &newCategory, IsActive: &falseValue})
	if product.Name != "New name" || product.Slug != "slug" || product.Description != "Description" ||
		product.SeoTitle != "Title" || product.SeoDescription != "SEO" || *product.CategoryID != newCategory || product.IsActive {
		t.Fatalf("product after update = %+v", product)
	}
	product.UpdateState(ProductChanges{})
	if product.Name != "New name" || product.IsActive {
		t.Fatalf("omitted fields changed product: %+v", product)
	}
}

func TestProductUpdateStateLeavesExplicitEmptyTextForValidation(t *testing.T) {
	category := 2
	product := Product{Name: "Name", Slug: "slug", Description: "Description", SeoTitle: "Title", SeoDescription: "SEO", CategoryID: &category}
	empty := " "
	product.UpdateState(ProductChanges{Name: &empty})
	if err := product.Validate(); err != ErrProductNameRequired {
		t.Fatalf("validation error = %v, want %v", err, ErrProductNameRequired)
	}
}

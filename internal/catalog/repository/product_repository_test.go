package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"e-commerce-go/internal/catalog/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestProductRepositoryListFiltersTotalAndPagination(t *testing.T) {
	databaseURL := os.Getenv("PRODUCT_REPOSITORY_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PRODUCT_REPOSITORY_TEST_DATABASE_URL is required")
	}
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}

	tests := []struct {
		name          string
		limit, offset int
		filters       func(int) domain.ProductListFilters
		want          []string
		total         int64
	}{
		{"unfiltered pagination", 2, 1, func(int) domain.ProductListFilters { return domain.ProductListFilters{} }, []string{"Inactive other", "Inactive selected"}, 4},
		{"category", 10, 0, func(id int) domain.ProductListFilters { return domain.ProductListFilters{CategoryID: &id} }, []string{"Inactive selected", "Active selected"}, 2},
		{"active true", 10, 0, func(int) domain.ProductListFilters { return domain.ProductListFilters{IsActive: boolPointer(true)} }, []string{"Active other", "Active selected"}, 2},
		{"active false", 10, 0, func(int) domain.ProductListFilters { return domain.ProductListFilters{IsActive: boolPointer(false)} }, []string{"Inactive other", "Inactive selected"}, 2},
		{"combined", 10, 0, func(id int) domain.ProductListFilters {
			return domain.ProductListFilters{CategoryID: &id, IsActive: boolPointer(false)}
		}, []string{"Inactive selected"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.Begin()
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}
			defer tx.Rollback()
			categoryID := insertProductFixtures(t, tx)
			products, total, err := NewProductRepository(tx).List(context.Background(), tt.limit, tt.offset, tt.filters(categoryID))
			if err != nil {
				t.Fatal(err)
			}
			if total != tt.total {
				t.Errorf("total=%d want=%d", total, tt.total)
			}
			if len(products) != len(tt.want) {
				t.Fatalf("count=%d want=%d", len(products), len(tt.want))
			}
			for i, want := range tt.want {
				if products[i].Name != want {
					t.Errorf("product[%d]=%q want=%q", i, products[i].Name, want)
				}
			}
		})
	}
}

func insertProductFixtures(t *testing.T, tx *gorm.DB) int {
	t.Helper()
	suffix := time.Now().UTC().Format("20060102150405.000000000")
	selected := domain.Category{Name: "Selected", Slug: "selected-" + suffix, Description: "Selected", IsActive: true}
	other := domain.Category{Name: "Other", Slug: "other-" + suffix, Description: "Other", IsActive: true}
	for _, category := range []*domain.Category{&selected, &other} {
		if err := tx.Create(category).Error; err != nil {
			t.Fatal(err)
		}
	}
	fixtures := []domain.Product{
		{Name: "Active selected", Slug: "active-selected-" + suffix, CategoryID: &selected.ID, IsActive: true},
		{Name: "Inactive selected", Slug: "inactive-selected-" + suffix, CategoryID: &selected.ID, IsActive: false},
		{Name: "Inactive other", Slug: "inactive-other-" + suffix, CategoryID: &other.ID, IsActive: false},
		{Name: "Active other", Slug: "active-other-" + suffix, CategoryID: &other.ID, IsActive: true},
	}
	for i := range fixtures {
		active := fixtures[i].IsActive
		if err := tx.Create(&fixtures[i]).Error; err != nil {
			t.Fatal(err)
		}
		if !active {
			if err := tx.Model(&fixtures[i]).UpdateColumn("is_active", false).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	return selected.ID
}

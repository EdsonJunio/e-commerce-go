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

func TestCategoryRepositoryListFiltersTotalAndPagination(t *testing.T) {
	databaseURL := os.Getenv("CATEGORY_REPOSITORY_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("CATEGORY_REPOSITORY_TEST_DATABASE_URL is required for the PostgreSQL repository test")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL test database: %v", err)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		filters   func(parentID int) domain.CategoryListFilters
		wantNames []string
		wantTotal int64
	}{
		{
			name:      "without filters preserves pagination",
			limit:     2,
			offset:    1,
			filters:   func(int) domain.CategoryListFilters { return domain.CategoryListFilters{} },
			wantNames: []string{"Inactive child", "Active child"},
			wantTotal: 4,
		},
		{
			name:  "parent filter",
			limit: 10,
			filters: func(parentID int) domain.CategoryListFilters {
				return domain.CategoryListFilters{ParentID: intPointer(parentID)}
			},
			wantNames: []string{"Inactive child", "Active child"},
			wantTotal: 2,
		},
		{
			name:      "active false filter",
			limit:     10,
			filters:   func(int) domain.CategoryListFilters { return domain.CategoryListFilters{IsActive: boolPointer(false)} },
			wantNames: []string{"Inactive root", "Inactive child"},
			wantTotal: 2,
		},
		{
			name:  "combined filters",
			limit: 10,
			filters: func(parentID int) domain.CategoryListFilters {
				return domain.CategoryListFilters{ParentID: intPointer(parentID), IsActive: boolPointer(false)}
			},
			wantNames: []string{"Inactive child"},
			wantTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.Begin()
			if tx.Error != nil {
				t.Fatalf("begin transaction: %v", tx.Error)
			}
			defer tx.Rollback()

			parentID := insertCategoryFixtures(t, tx)
			repository := NewCategoryRepository(tx, nil)
			categories, total, err := repository.List(context.Background(), tt.limit, tt.offset, tt.filters(parentID))
			if err != nil {
				t.Fatalf("list categories: %v", err)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if len(categories) != len(tt.wantNames) {
				t.Fatalf("category count = %d, want %d", len(categories), len(tt.wantNames))
			}
			for index, wantName := range tt.wantNames {
				if categories[index].Name != wantName {
					t.Errorf("category[%d].Name = %q, want %q", index, categories[index].Name, wantName)
				}
			}
		})
	}
}

func insertCategoryFixtures(t *testing.T, tx *gorm.DB) int {
	t.Helper()
	suffix := time.Now().UTC().Format("20060102150405.000000000")
	parent := domain.Category{Name: "Parent", Slug: "parent-" + suffix, Description: "Parent", IsActive: true}
	if err := tx.Create(&parent).Error; err != nil {
		t.Fatalf("insert parent category: %v", err)
	}

	fixtures := []domain.Category{
		{Name: "Active child", Slug: "active-child-" + suffix, ParentID: &parent.ID, IsActive: true, Description: "Active child"},
		{Name: "Inactive child", Slug: "inactive-child-" + suffix, ParentID: &parent.ID, IsActive: false, Description: "Inactive child"},
		{Name: "Inactive root", Slug: "inactive-root-" + suffix, IsActive: false, Description: "Inactive root"},
	}
	for index := range fixtures {
		isActive := fixtures[index].IsActive
		if err := tx.Create(&fixtures[index]).Error; err != nil {
			t.Fatalf("insert category fixture %q: %v", fixtures[index].Name, err)
		}
		if !isActive {
			if err := tx.Model(&fixtures[index]).UpdateColumn("is_active", false).Error; err != nil {
				t.Fatalf("mark category fixture %q inactive: %v", fixtures[index].Name, err)
			}
		}
	}
	return parent.ID
}

func intPointer(value int) *int    { return &value }
func boolPointer(value bool) *bool { return &value }

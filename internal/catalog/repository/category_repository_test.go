package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/cache"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestCategoryRepositoryUpdateRejectsCycles(t *testing.T) {
	db := categoryTestDatabase(t)
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	defer tx.Rollback()
	root := domain.Category{Name: "Cycle root", Slug: uniqueCategorySlug("cycle-root"), Description: "Root", IsActive: true}
	child := domain.Category{Name: "Cycle child", Slug: uniqueCategorySlug("cycle-child"), Description: "Child", IsActive: true}
	grandchild := domain.Category{Name: "Cycle grandchild", Slug: uniqueCategorySlug("cycle-grandchild"), Description: "Grandchild", IsActive: true}
	for _, category := range []*domain.Category{&root, &child, &grandchild} {
		if err := tx.Create(category).Error; err != nil {
			t.Fatalf("insert category: %v", err)
		}
	}
	repository := NewCategoryRepository(tx, &cache.RedisClient{Client: redisClient})
	child.ParentID = &root.ID
	grandchild.ParentID = &child.ID
	for _, category := range []*domain.Category{&child, &grandchild} {
		if err := repository.Update(context.Background(), category); err != nil {
			t.Fatalf("assign valid parent: %v", err)
		}
	}

	root.ParentID = &grandchild.ID
	if err := repository.Update(context.Background(), &root); !errors.Is(err, domain.ErrInvalidCategoryReference) {
		t.Fatalf("indirect cycle error = %v", err)
	}
	root.ParentID = &root.ID
	if err := repository.Update(context.Background(), &root); !errors.Is(err, domain.ErrInvalidCategoryReference) {
		t.Fatalf("direct cycle error = %v", err)
	}
	var storedRoot domain.Category
	if err := tx.First(&storedRoot, root.ID).Error; err != nil || storedRoot.ParentID != nil {
		t.Fatalf("root after rejected updates = %+v, error = %v", storedRoot, err)
	}
	grandchild.ParentID = nil
	if err := repository.Update(context.Background(), &grandchild); err != nil {
		t.Fatalf("clear parent: %v", err)
	}
	var storedGrandchild domain.Category
	if err := tx.First(&storedGrandchild, grandchild.ID).Error; err != nil || storedGrandchild.ParentID != nil {
		t.Fatalf("grandchild after clear = %+v, error = %v", storedGrandchild, err)
	}
	staleKey := fmt.Sprintf("category:id:%d", grandchild.ID)
	if err := redisClient.Set(context.Background(), staleKey, fmt.Sprintf(`{"parent_id":%d}`, child.ID), 0).Err(); err != nil {
		t.Fatalf("seed stale cache entry: %v", err)
	}
	parentID, err := repository.FindParentByID(context.Background(), grandchild.ID)
	if err != nil || parentID != nil {
		t.Fatalf("authoritative parent after clear = %v, error = %v", parentID, err)
	}
}

func TestCategoryRepositoryConcurrentParentUpdatesDoNotCycle(t *testing.T) {
	db := categoryTestDatabase(t)
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	first := domain.Category{Name: "Concurrent first", Slug: uniqueCategorySlug("concurrent-first"), Description: "First", IsActive: true}
	second := domain.Category{Name: "Concurrent second", Slug: uniqueCategorySlug("concurrent-second"), Description: "Second", IsActive: true}
	for _, category := range []*domain.Category{&first, &second} {
		if err := db.Create(category).Error; err != nil {
			t.Fatalf("insert category: %v", err)
		}
	}
	t.Cleanup(func() {
		if err := db.Model(&domain.Category{}).Where("id IN ?", []int{first.ID, second.ID}).Update("deleted_at", time.Now()).Error; err != nil {
			t.Errorf("hide concurrent test categories: %v", err)
		}
	})
	repository := NewCategoryRepository(db, &cache.RedisClient{Client: redisClient})
	first.ParentID = &second.ID
	second.ParentID = &first.ID
	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for _, category := range []*domain.Category{&first, &second} {
		workers.Add(1)
		go func(category *domain.Category) {
			defer workers.Done()
			<-start
			results <- repository.Update(context.Background(), category)
		}(category)
	}
	close(start)
	workers.Wait()
	close(results)
	successes, rejected := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrInvalidCategoryReference):
			rejected++
		default:
			t.Fatalf("unexpected update error: %v", err)
		}
	}
	if successes != 1 || rejected != 1 {
		t.Fatalf("successes = %d, rejected = %d", successes, rejected)
	}
}

func categoryTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	databaseURL := os.Getenv("CATEGORY_REPOSITORY_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("CATEGORY_REPOSITORY_TEST_DATABASE_URL is required for the PostgreSQL repository test")
	}
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL test database: %v", err)
	}
	return db
}

func uniqueCategorySlug(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("20060102150405.000000000")
}

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

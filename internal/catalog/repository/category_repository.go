package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/cache"
	"e-commerce-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db    *gorm.DB
	cache *cache.RedisClient
}

// NewCategoryRepository returns a new category repository backed by GORM and Redis.
func NewCategoryRepository(db *gorm.DB, cache *cache.RedisClient) domain.CategoryRepository {
	return &categoryRepository{
		db:    db,
		cache: cache,
	}
}

func (r *categoryRepository) List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]domain.Category, int64, error) {
	var categories []domain.Category
	var total int64

	tx := r.db.WithContext(ctx).Model(&domain.Category{})

	if name, ok := filters["name"]; ok {
		tx = tx.Where("name = ?", name)
	}
	if active, ok := filters["is_active"]; ok {
		tx = tx.Where("is_active = ?", active)
	}
	if parentID, ok := filters["parent_id"]; ok {
		tx = tx.Where("parent_id = ?", parentID)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tx = tx.Order("id DESC")

	if err := tx.Offset(offset).Limit(limit).Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id int) (*domain.Category, error) {
	// 1. Build cache key (Example: "category:id:15")
	cacheKey := fmt.Sprintf("category:id:%d", id)

	// 2. Try Redis first
	val, err := r.cache.Client.Get(ctx, cacheKey).Result()
	if err == nil {
		// CACHE HIT
		var category domain.Category
		if jsonErr := json.Unmarshal([]byte(val), &category); jsonErr == nil {
			logger.L().Info("cache hit: returning category from redis", zap.String("key", cacheKey))
			return &category, nil
		}
	}

	// 3. CACHE MISS → Fetch from database
	var category domain.Category
	err = r.db.WithContext(ctx).First(&category, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}

	// 4. Save to Redis with TTL (10 minutes)
	if data, jsonErr := json.Marshal(category); jsonErr == nil {
		_ = r.cache.Client.Set(ctx, cacheKey, data, 10*time.Minute).Err()
	}

	return &category, nil
}

func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	// 1. Build cache key (Example: "category:slug:electronics")
	cacheKey := fmt.Sprintf("category:slug:%s", slug)

	// 2. Try Redis
	val, err := r.cache.Client.Get(ctx, cacheKey).Result()
	if err == nil {
		var category domain.Category
		if jsonErr := json.Unmarshal([]byte(val), &category); jsonErr == nil {
			logger.L().Info("cache hit: returning category from redis", zap.String("key", cacheKey))
			return &category, nil
		}
	}

	// 3. Fetch from database
	var category domain.Category
	err = r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}

	// 4. Save to Redis (10 minutes)
	if data, jsonErr := json.Marshal(category); jsonErr == nil {
		_ = r.cache.Client.Set(ctx, cacheKey, data, 10*time.Minute).Err()
	}

	return &category, nil
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrCategorySlugExists
		}
		return err
	}
	return nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	err := r.db.WithContext(ctx).Save(category).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrCategorySlugExists
		}
		return err
	}

	// Cache invalidation — since data changed, remove old cached entries
	_ = r.cache.Client.Del(ctx, fmt.Sprintf("category:id:%d", category.ID))
	_ = r.cache.Client.Del(ctx, fmt.Sprintf("category:slug:%s", category.Slug))

	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	res := r.db.WithContext(ctx).Delete(&domain.Category{}, id)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrCategoryNotFound
	}

	// Cache invalidation — remove cached version by ID
	_ = r.cache.Client.Del(ctx, fmt.Sprintf("category:id:%d", id))

	// Note: To remove slug cache as well, you would need to fetch the record before deleting it.

	return nil
}

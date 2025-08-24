package repository

import (
	"errors"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/pkg/logger"

	"github.com/jackc/pgconn"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository returns a new product repository backed by GORM.
// The caller is responsible for managing the lifecycle of db.
func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

// List returns products with pagination and filters.
func (r *productRepository) List(limit, offset int, filters map[string]interface{}) ([]domain.Product, int64, error) {
	var products []domain.Product
	var total int64

	tx := r.db.Model(&domain.Product{})
	for key, value := range filters {
		tx = tx.Where(key, value)
	}

	if err := tx.Count(&total).Error; err != nil {
		logger.L().Error(
			"failed to count products",
			zap.Error(err),
			zap.Any("filters", filters),
		)
		return nil, 0, err
	}

	if err := tx.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		logger.L().Error(
			"failed to list products",
			zap.Error(err),
			zap.Any("filters", filters),
		)
		return nil, 0, err
	}

	return products, total, nil
}

// FindByID fetches a product by ID.
// Returns ErrNotFound if no record exists.
func (r *productRepository) FindByID(id int) (*domain.Product, error) {
	var product domain.Product
	err := r.db.First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		logger.L().Info("product not found by ID", zap.Int("id", id))
		return nil, domain.ErrNotFound
	}
	if err != nil {
		logger.L().Error(
			"failed to fetch product by ID",
			zap.Int("id", id),
			zap.Error(err),
		)
		return nil, err
	}
	return &product, nil
}

// FindBySlug fetches a product by slug.
// Returns (nil, nil) if no record exists.
func (r *productRepository) FindBySlug(slug string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.Where("slug = ?", slug).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		logger.L().Info("product not found by slug", zap.String("slug", slug))
		return nil, nil
	}
	if err != nil {
		logger.L().Error(
			"failed to fetch product by slug",
			zap.String("slug", slug),
			zap.Error(err),
		)
		return nil, err
	}
	return &product, nil
}

// Create inserts a new product in the database.
// Returns ErrSlugExists if the slug already exists.
func (r *productRepository) Create(product *domain.Product) error {
	err := r.db.Create(product).Error
	if err != nil {
		if isUniqueViolation(err) {
			logger.L().Warn(
				"unique constraint violation while creating product",
				zap.String("slug", product.Slug),
				zap.Error(err),
			)
			return domain.ErrSlugExists
		}
		logger.L().Error(
			"failed to create product in database",
			zap.String("slug", product.Slug),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// Update saves product changes in the database.
// Returns ErrSlugExists if another record already uses the same slug.
func (r *productRepository) Update(product *domain.Product) error {
	err := r.db.Save(product).Error
	if err != nil {
		if isUniqueViolation(err) {
			logger.L().Warn(
				"unique constraint violation while updating product",
				zap.Int("id", product.ID),
				zap.String("slug", product.Slug),
				zap.Error(err),
			)
			return domain.ErrSlugExists
		}
		logger.L().Error(
			"failed to update product",
			zap.Int("id", product.ID),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// Delete removes a product by ID.
// Returns ErrNotFound if no record was affected.
func (r *productRepository) Delete(id int) error {
	res := r.db.Delete(&domain.Product{}, id)
	if res.Error != nil {
		logger.L().Error(
			"failed to delete product",
			zap.Int("id", id),
			zap.Error(res.Error),
		)
		return res.Error
	}
	if res.RowsAffected == 0 {
		logger.L().Info("product not found to delete", zap.Int("id", id))
		return domain.ErrNotFound
	}
	return nil
}

// isUniqueViolation checks if the error corresponds to a UNIQUE constraint violation.
func isUniqueViolation(err error) bool {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) {
		return pgxErr.Code == "23505"
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == "23505"
	}
	return false
}

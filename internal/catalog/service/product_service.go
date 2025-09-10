package service

import (
	"errors"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/pkg/logger"

	"go.uber.org/zap"
)

type productService struct {
	repo domain.ProductRepository
}

// NewProductService returns a new instance of productService implementing domain.ProductService.
func NewProductService(repo domain.ProductRepository) domain.ProductService {
	return &productService{repo: repo}
}

// ListProducts returns a paginated list of products with optional filters.
func (s *productService) ListProducts(p domain.Pagination, filters map[string]interface{}) ([]domain.Product, int64, error) {
	logger.L().Info(
		"listing products",
		zap.Int("limit", p.Limit),
		zap.Int("page", p.Page),
		zap.Int("offset", p.Offset),
		zap.Any("filters", filters),
	)

	return s.repo.List(p.Limit, p.Offset, filters)
}

// GetProductByID retrieves a product by ID.
func (s *productService) GetProductByID(id int) (*domain.Product, error) {
	if id <= 0 {
		logger.L().Warn("invalid product ID in GetProductByID", zap.Int("id", id))
		return nil, domain.ErrInvalidID
	}

	product, err := s.repo.FindByID(id)
	if err != nil {
		logger.L().Error(
			"failed to fetch product by ID",
			zap.Int("id", id),
			zap.Error(err),
		)
		return nil, err
	}

	if product == nil {
		logger.L().Info("product not found by ID", zap.Int("id", id))
		return nil, domain.ErrNotFound
	}

	return product, nil
}

// GetProductBySlug retrieves a product by slug.
func (s *productService) GetProductBySlug(slug string) (*domain.Product, error) {
	if slug == "" {
		logger.L().Warn("empty slug in GetProductBySlug")
		return nil, domain.ErrSlugIsReq
	}

	product, err := s.repo.FindBySlug(slug)
	if err != nil {
		logger.L().Error(
			"failed to fetch product by slug",
			zap.String("slug", slug),
			zap.Error(err),
		)
		return nil, err
	}

	if product == nil {
		logger.L().Info("product not found by slug", zap.String("slug", slug))
		return nil, domain.ErrNotFound
	}

	return product, nil
}

// CreateProduct validates and persists a new product.
func (s *productService) CreateProduct(product *domain.Product) error {
	if product.Name == "" {
		logger.L().Warn("missing product name in CreateProduct")
		return domain.ErrNameReq
	}
	if product.Slug == "" {
		logger.L().Warn("missing product slug in CreateProduct")
		return domain.ErrSlugReq
	}
	if product.Description == "" {
		logger.L().Warn(
			"invalid product description in CreateProduct",
			zap.String("description", product.Description),
		)
		return domain.ErrPriceInvalid
	}

	existing, err := s.repo.FindBySlug(product.Slug)
	if err != nil {
		logger.L().Error(
			"failed to fetch product by slug",
			zap.String("slug", product.Slug),
			zap.Error(err),
		)
		return err
	}
	if existing != nil {
		logger.L().Warn("duplicate product slug", zap.String("slug", product.Slug))
		return domain.ErrSlugExists
	}

	logger.L().Info("creating new product", zap.String("slug", product.Slug))
	return s.repo.Create(product)
}

// UpdateProduct updates an existing product.
func (s *productService) UpdateProduct(id int, product *domain.Product) error {
	if id <= 0 {
		logger.L().Warn("invalid product ID in UpdateProduct", zap.Int("id", id))
		return domain.ErrInvalidID
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		logger.L().Error(
			"failed to fetch existing product",
			zap.Int("id", id),
			zap.Error(err),
		)
		return err
	}
	if existing == nil {
		logger.L().Info("product not found for update", zap.Int("id", id))
		return domain.ErrNotFound
	}

	existing.Name = product.Name
	existing.Slug = product.Slug
	existing.Description = product.Description
	existing.IsActive = product.IsActive
	existing.CategoryID = product.CategoryID

	logger.L().Info("updating product", zap.Int("id", id))
	return s.repo.Update(existing)
}

// DeleteProduct deletes a product by ID.
func (s *productService) DeleteProduct(id int) error {
	if id <= 0 {
		logger.L().Warn("invalid product ID in DeleteProduct", zap.Int("id", id))
		return domain.ErrInvalidID
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		logger.L().Error(
			"failed to fetch product for deletion",
			zap.Int("id", id),
			zap.Error(err),
		)
		return err
	}
	if existing == nil {
		logger.L().Info("product not found for deletion", zap.Int("id", id))
		return domain.ErrNotFound
	}

	logger.L().Info("deleting product", zap.Int("id", id))
	return s.repo.Delete(id)
}

var _ = errors.Is

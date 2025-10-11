package service

import (
	"errors"

	"e-commerce-go/internal/catalog/domain"
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
	return s.repo.List(p.Limit, p.Offset, filters)
}

// GetProductByID retrieves a product by ID.
func (s *productService) GetProductByID(id int) (*domain.Product, error) {
	if id <= 0 {
		return nil, domain.ErrProductNotFound
	}

	product, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, domain.ErrProductNotFound
	}

	return product, nil
}

// GetProductBySlug retrieves a product by slug.
func (s *productService) GetProductBySlug(slug string) (*domain.Product, error) {
	if slug == "" {
		return nil, domain.ErrProductDescriptionRequired
	}

	product, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, domain.ErrProductNotFound
	}

	return product, nil
}

// CreateProduct validates and persists a new product.
func (s *productService) CreateProduct(product *domain.Product) error {
	if product.Name == "" {
		return domain.ErrProductNameRequired
	}
	if product.Slug == "" {
		return domain.ErrProductSlugRequired
	}
	if product.Description == "" {
		return domain.ErrProductDescriptionRequired
	}

	existing, err := s.repo.FindBySlug(product.Slug)
	if err != nil {
		return err
	}
	if existing != nil {
		return domain.ErrProductSlugExists
	}

	_, err = s.repo.FindByID(*product.CategoryID)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return domain.ErrInvalidCategoryReference
		}
		return err
	}

	return s.repo.Create(product)
}

// UpdateProduct updates an existing product.
func (s *productService) UpdateProduct(id int, product *domain.Product) error {
	if id <= 0 {
		return domain.ErrInvalidProductID
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrProductNotFound
	}

	existing.Name = product.Name
	existing.Slug = product.Slug
	existing.Description = product.Description
	existing.IsActive = product.IsActive
	existing.CategoryID = product.CategoryID

	return s.repo.Update(existing)
}

// DeleteProduct deletes a product by ID.
func (s *productService) DeleteProduct(id int) error {
	if id <= 0 {
		return domain.ErrInvalidProductID
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrProductNotFound
	}

	return s.repo.Delete(id)
}

var _ = errors.Is

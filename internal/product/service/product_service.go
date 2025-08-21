package service

import (
	"e-commerce-go/internal/product/domain"
	"errors"
)

type productService struct {
	repo domain.Repository
}

// NewProductService cria uma nova instância do serviço de produtos
func NewProductService(repo domain.Repository) domain.Service {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(product *domain.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.Slug == "" {
		return errors.New("product slug is required")
	}
	if product.PriceCents <= 0 {
		return errors.New("product price must be greater than zero")
	}

	existing, _ := s.repo.FindBySlug(product.Slug)
	if existing != nil {
		return errors.New("product with this slug already exists")
	}

	return s.repo.Create(product)
}

func (s *productService) GetProductByID(id int) (*domain.Product, error) {
	if id <= 0 {
		return nil, errors.New("invalid product ID")
	}

	product, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, errors.New("product not found")
	}

	return product, nil
}

func (s *productService) GetProductBySlug(slug string) (*domain.Product, error) {
	if slug == "" {
		return nil, errors.New("slug is required")
	}

	product, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, errors.New("product not found")
	}

	return product, nil
}

func (s *productService) UpdateProduct(id int, product *domain.Product) error {
	if id <= 0 {
		return errors.New("invalid product ID")
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if existing == nil {
		return errors.New("product not found")
	}

	existing.Name = product.Name
	existing.Slug = product.Slug
	existing.PriceCents = product.PriceCents
	existing.IsActive = product.IsActive
	existing.CategoryID = product.CategoryID

	return s.repo.Update(existing)
}

func (s *productService) DeleteProduct(id int) error {
	if id <= 0 {
		return errors.New("invalid product ID")
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if existing == nil {
		return errors.New("product not found")
	}

	return s.repo.Delete(id)
}

func (s *productService) ListProducts(limit, page int, filters map[string]interface{}) ([]domain.Product, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	return s.repo.List(limit, offset, filters)
}

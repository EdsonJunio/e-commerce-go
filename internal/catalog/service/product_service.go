package service

import (
	"errors"

	"e-commerce-go/internal/catalog/domain"
)

type productService struct {
	repo domain.ProductRepository
	cate domain.CategoryRepository
}

// NewProductService returns a new instance of productService implementing domain.ProductService.
func NewProductService(repo domain.ProductRepository, cate domain.CategoryRepository) domain.ProductService {
	return &productService{repo: repo, cate: cate}
}

func (s *productService) ListProducts(p domain.Pagination, filters map[string]interface{}) ([]domain.Product, int64, error) {
	return s.repo.List(p.Limit, p.Offset, filters)
}

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
	if product.SeoTitle == "" {
		return domain.ErrSeoTitle
	}
	if product.SeoDescription == "" {
		return domain.ErrSeoDescription
	}

	_, err := s.cate.FindByID(*product.CategoryID)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return domain.ErrInvalidCategoryReference
		}
		return err
	}

	return s.repo.Create(product)
}

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

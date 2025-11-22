package service

import (
	"context"
	"errors"

	"e-commerce-go/internal/catalog/domain"
)

type productService struct {
	repo domain.ProductRepository
	cate domain.CategoryRepository
}

// NewProductService returns a new instance of productService.
func NewProductService(
	repo domain.ProductRepository,
	cate domain.CategoryRepository) domain.ProductService {
	return &productService{repo: repo, cate: cate}
}

func (s *productService) ListProducts(ctx context.Context, p domain.Pagination, filters map[string]interface{}) ([]domain.Product, int64, error) {
	return s.repo.List(ctx, p.Limit, p.Offset, filters)
}

func (s *productService) GetProductByID(ctx context.Context, id int) (*domain.Product, error) {
	return s.findProductOrFail(ctx, id)
}

func (s *productService) GetProductBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	if slug == "" {
		return nil, domain.ErrProductSlugRequired
	}

	product, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, domain.ErrProductNotFound
	}

	return product, nil
}

func (s *productService) CreateProduct(ctx context.Context, product *domain.Product) error {
	if err := product.Validate(); err != nil {
		return err
	}

	if err := s.ensureCategoryExists(ctx, *product.CategoryID); err != nil {
		return err
	}

	return s.repo.Create(ctx, product)
}

func (s *productService) UpdateProduct(ctx context.Context, id int, req *domain.Product) error {
	existing, err := s.findProductOrFail(ctx, id)
	if err != nil {
		return err
	}

	existing.UpdateState(req)

	if err := existing.Validate(); err != nil {
		return err
	}

	if req.CategoryID != nil {
		if err := s.ensureCategoryExists(ctx, *req.CategoryID); err != nil {
			return err
		}
	}

	return s.repo.Update(ctx, existing)
}

func (s *productService) DeleteProduct(ctx context.Context, id int) error {
	if _, err := s.findProductOrFail(ctx, id); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *productService) findProductOrFail(ctx context.Context, id int) (*domain.Product, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidProductID
	}

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, domain.ErrProductNotFound
	}

	return product, nil
}

func (s *productService) ensureCategoryExists(ctx context.Context, categoryID int) error {
	_, err := s.cate.FindByID(categoryID)

	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return domain.ErrInvalidCategoryReference
		}
		return err
	}

	return nil
}

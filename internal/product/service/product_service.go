package service

import (
	"e-commerce-go/internal/product/domain"
	"e-commerce-go/pkg/logger"
	"errors"

	"go.uber.org/zap"
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
		logger.L().Warn("Nome do produto ausente em CreateProduct")
		return errors.New("product name is required")
	}
	if product.Slug == "" {
		logger.L().Warn("Slug do produto ausente em CreateProduct")
		return errors.New("product slug is required")
	}
	if product.PriceCents <= 0 {
		logger.L().Warn("Preço inválido em CreateProduct", zap.Int64("price_cents", product.PriceCents))
		return errors.New("product price must be greater than zero")
	}

	existing, err := s.repo.FindBySlug(product.Slug)
	if err != nil {
		logger.L().Error("Erro ao buscar produto por slug", zap.String("slug", product.Slug), zap.Error(err))
		return err
	}
	if existing != nil {
		logger.L().Warn("Produto com slug duplicado", zap.String("slug", product.Slug))
		return errors.New("product with this slug already exists")
	}

	logger.L().Info("Criando novo produto", zap.String("slug", product.Slug))
	return s.repo.Create(product)
}

func (s *productService) GetProductByID(id int) (*domain.Product, error) {
	if id <= 0 {
		logger.L().Warn("ID inválido em GetProductByID", zap.Int("id", id))
		return nil, errors.New("invalid product ID")
	}

	product, err := s.repo.FindByID(id)
	if err != nil {
		logger.L().Error("Erro ao buscar produto por ID", zap.Int("id", id), zap.Error(err))
		return nil, err
	}

	if product == nil {
		logger.L().Info("Produto não encontrado por ID", zap.Int("id", id))
		return nil, errors.New("product not found")
	}

	return product, nil
}

func (s *productService) GetProductBySlug(slug string) (*domain.Product, error) {
	if slug == "" {
		logger.L().Warn("Slug vazio em GetProductBySlug")
		return nil, errors.New("slug is required")
	}

	product, err := s.repo.FindBySlug(slug)
	if err != nil {
		logger.L().Error("Erro ao buscar produto por slug", zap.String("slug", slug), zap.Error(err))
		return nil, err
	}

	if product == nil {
		logger.L().Info("Produto não encontrado por slug", zap.String("slug", slug))
		return nil, errors.New("product not found")
	}

	return product, nil
}

func (s *productService) UpdateProduct(id int, product *domain.Product) error {
	if id <= 0 {
		logger.L().Warn("ID inválido em UpdateProduct", zap.Int("id", id))
		return errors.New("invalid product ID")
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		logger.L().Error("Erro ao buscar produto existente", zap.Int("id", id), zap.Error(err))
		return err
	}

	if existing == nil {
		logger.L().Info("Produto não encontrado para atualização", zap.Int("id", id))
		return errors.New("product not found")
	}

	existing.Name = product.Name
	existing.Slug = product.Slug
	existing.PriceCents = product.PriceCents
	existing.IsActive = product.IsActive
	existing.CategoryID = product.CategoryID

	logger.L().Info("Atualizando produto", zap.Int("id", id))
	return s.repo.Update(existing)
}

func (s *productService) DeleteProduct(id int) error {
	if id <= 0 {
		logger.L().Warn("ID inválido em DeleteProduct", zap.Int("id", id))
		return errors.New("invalid product ID")
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		logger.L().Error("Erro ao buscar produto para deletar", zap.Int("id", id), zap.Error(err))
		return err
	}

	if existing == nil {
		logger.L().Info("Produto não encontrado para deletar", zap.Int("id", id))
		return errors.New("product not found")
	}

	logger.L().Info("Deletando produto", zap.Int("id", id))
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

	logger.L().Info("Listando produtos",
		zap.Int("limit", limit),
		zap.Int("page", page),
		zap.Int("offset", offset),
		zap.Any("filters", filters),
	)

	return s.repo.List(limit, offset, filters)
}

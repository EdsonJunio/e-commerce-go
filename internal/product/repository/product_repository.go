package repository

import (
	"e-commerce-go/internal/product/domain"
	"e-commerce-go/pkg/logger"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository cria uma nova instância do repositório de produtos
func NewProductRepository(db *gorm.DB) domain.Repository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *domain.Product) error {
	err := r.db.Create(product).Error
	if err != nil {
		logger.L().Error("Erro ao criar produto no banco", zap.Error(err), zap.String("slug", product.Slug))
	}
	return err
}

func (r *productRepository) FindByID(id int) (*domain.Product, error) {
	var product domain.Product
	err := r.db.First(&product, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		logger.L().Info("Produto não encontrado por ID", zap.Int("id", id))
		return nil, nil
	}
	if err != nil {
		logger.L().Error("Erro ao buscar produto por ID", zap.Int("id", id), zap.Error(err))
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) FindBySlug(slug string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.Where("slug = ?", slug).First(&product).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		logger.L().Info("Produto não encontrado por slug", zap.String("slug", slug))
		return nil, nil
	}
	if err != nil {
		logger.L().Error("Erro ao buscar produto por slug", zap.String("slug", slug), zap.Error(err))
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) Update(product *domain.Product) error {
	err := r.db.Save(product).Error
	if err != nil {
		logger.L().Error("Erro ao atualizar produto", zap.Int("id", product.ID), zap.Error(err))
	}
	return err
}

func (r *productRepository) Delete(id int) error {
	err := r.db.Delete(&domain.Product{}, id).Error
	if err != nil {
		logger.L().Error("Erro ao deletar produto", zap.Int("id", id), zap.Error(err))
	}
	return err
}

func (r *productRepository) List(limit, offset int, filters map[string]interface{}) ([]domain.Product, int64, error) {
	var products []domain.Product
	var total int64

	tx := r.db.Model(&domain.Product{})

	for key, value := range filters {
		tx = tx.Where(key, value)
	}

	if err := tx.Count(&total).Error; err != nil {
		logger.L().Error("Erro ao contar produtos", zap.Error(err), zap.Any("filters", filters))
		return nil, 0, err
	}

	err := tx.Offset(offset).Limit(limit).Find(&products).Error
	if err != nil {
		logger.L().Error("Erro ao listar produtos", zap.Error(err), zap.Any("filters", filters))
		return nil, 0, err
	}

	return products, total, nil
}

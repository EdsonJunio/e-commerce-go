package repository

import (
	"errors"
	"gorm.io/gorm"
	"e-commerce-go/internal/product/domain"
)

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository cria uma nova instância do repositório de produtos
func NewProductRepository(db *gorm.DB) domain.Repository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *domain.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) FindByID(id int) (*domain.Product, error) {
	var product domain.Product
	err := r.db.First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &product, err
}

func (r *productRepository) FindBySlug(slug string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.Where("slug = ?", slug).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &product, err
}

func (r *productRepository) Update(product *domain.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id int) error {
	return r.db.Delete(&domain.Product{}, id).Error
}

func (r *productRepository) List(limit, offset int, filters map[string]interface{}) ([]domain.Product, int64, error) {
	var products []domain.Product
	var total int64

	tx := r.db.Model(&domain.Product{})

	// Aplicar filtros
	for key, value := range filters {
		tx = tx.Where(key, value)
	}

	// Contar o total de registros com os filtros aplicados
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Aplicar paginação e buscar resultados
	err := tx.Offset(offset).Limit(limit).Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

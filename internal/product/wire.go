// +build wireinject

package product

import (
	"e-commerce-go/internal/product/delivery/http"
	"e-commerce-go/internal/product/domain"
	"e-commerce-go/internal/product/repository"
	"e-commerce-go/internal/product/service"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// InitializeProductHandler inicializa o handler HTTP de produtos com todas as dependências
func InitializeProductHandler(db *gorm.DB) *http.ProductHandler {
	wire.Build(
		repository.NewProductRepository,
		service.NewProductService,
		http.NewProductHandler,
	)
	return &http.ProductHandler{}
}

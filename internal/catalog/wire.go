package catalog

import (
	"e-commerce-go/internal/catalog/delivery/http"
	"e-commerce-go/internal/catalog/repository"
	"e-commerce-go/internal/catalog/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// InitializeProductHandler Initialize Handler HTTP of Products with all dependencies
func InitializeProductHandler(db *gorm.DB) *http.ProductHandler {
	wire.Build(
		repository.NewProductRepository,
		service.NewProductService,
		http.NewProductHandler,
	)
	return &http.ProductHandler{}
}

package service

import (
	"e-commerce-go/internal/catalog/domain"
)

type productSKUService struct {
	repo domain.SkuRepository
}

func NewProductSKUService(
	repo domain.SkuRepository) domain.SkuService {
	return &productSKUService{repo: repo}
}

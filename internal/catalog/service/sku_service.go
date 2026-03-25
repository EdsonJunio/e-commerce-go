package service

import (
	"context"
	"e-commerce-go/internal/catalog/domain"
)

type ProductSkuService struct {
	repo domain.ProductSkusRepository
}

func NewProductSkuService(
	repo domain.ProductSkusRepository) domain.ProductSkuService {
	return &ProductSkuService{repo: repo}
}

func (ss *ProductSkuService) ListSkus(ctx context.Context, p domain.Pagination, filters map[string]interface{}) ([]domain.Product_skus, int64, error) {
	return ss.repo.List(ctx, p.Limit, p.Offset, filters)
}

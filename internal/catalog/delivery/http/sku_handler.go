package http

import (
	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/middleware"

	"github.com/gin-gonic/gin"
)

type SkuHandler struct {
	service domain.SkuService
}

func NewSkuHandler(service domain.SkuService) *SkuHandler {
	return &SkuHandler{service: service}
}

func (h *SkuHandler) RegisterSkuRoutes(router *gin.Engine, auth *middleware.AuthMiddleware) {

}

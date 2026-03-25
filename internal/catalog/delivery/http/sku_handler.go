package http

import (
	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/middleware"
	"e-commerce-go/internal/shared/response"
	"e-commerce-go/internal/shared/transport"
	"e-commerce-go/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductSkuHandler struct {
	service domain.ProductSkuService
}

func NewProductSkuHandler(service domain.ProductSkuService) *ProductSkuHandler {
	return &ProductSkuHandler{service: service}
}

func (sh *ProductSkuHandler) RegisterProductSkuRoutes(router *gin.Engine, auth *middleware.AuthMiddleware) {
	v1 := router.Group("/api/v1")
	{
		skus := v1.Group("/skus")
		{
			skus.GET("", sh.ListSkus)
		}
	}
}

func (sh *ProductSkuHandler) ListSkus(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	pagination := domain.NewPagination(page, limit)

	filters := make(map[string]interface{})
	if skusID := c.Query("sku_id"); skusID != "" {
		if skuID, err := strconv.Atoi(skusID); err == nil {
			filters["sku_id = ?"] = skuID
		}
	}

	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active = ?"] = active
		}
	}

	skus, total, err := sh.service.ListSkus(c.Request.Context(), pagination, filters)
	if err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to list skus",
			err,
			zap.Int("page", page),
			zap.Int("limit", limit),
			zap.String("request_id", reqID),
		)

		msg := err.Error()
		if mapping.HTTPCode == http.StatusInternalServerError {
			msg = "internal server error"
		}

		response.Error(c, mapping.HTTPCode, mapping.Code, msg)
		return
	}

	logger.L().Info("skus listed successfully",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("total", total),
		zap.String("request_id", reqID))

	response.SuccessPaginated(c, http.StatusOK, skus, total, page, limit)

}

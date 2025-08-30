package http

import (
	"net/http"
	"strconv"
	"strings"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/response"
	"e-commerce-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service domain.ProductService
}

type createProductRequest struct {
	CategoryID int    `json:"category_id" binding:"required,min=1"`
	Name       string `json:"name" binding:"required,min=3,max=255"`
	Slug       string `json:"slug" binding:"required,min=3,max=255"`
	PriceCents int64  `json:"price_cents" binding:"required,min=1"`
	IsActive   bool   `json:"is_active"`
}

type updateProductRequest struct {
	CategoryID *int    `json:"category_id,omitempty"`
	Name       *string `json:"name,omitempty"`
	Slug       *string `json:"slug,omitempty"`
	PriceCents *int64  `json:"price_cents,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
}

// NewProductHandler returns a new ProductHandler with the given service.
func NewProductHandler(service domain.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// RegisterProductRoutes registers product endpoints in the Gin router.
func (h *ProductHandler) RegisterProductRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.GET("", h.ListProducts)
			products.GET("/:id", h.GetProduct)
			products.GET("/slug/:slug", h.GetProductBySlug)
			products.POST("", h.CreateProduct)
			products.PUT("/:id", h.UpdateProduct)
			products.DELETE("/:id", h.DeleteProduct)
		}
	}
}

// ListProducts handles GET /products with pagination and filters.
func (h *ProductHandler) ListProducts(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	pagination := domain.NewPagination(page, limit)

	filters := make(map[string]interface{})
	if categoryID := c.Query("category_id"); categoryID != "" {
		if catID, err := strconv.Atoi(categoryID); err == nil {
			filters["category_id = ?"] = catID
		}
	}
	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active = ?"] = active
		}
	}

	products, total, err := h.service.ListProducts(pagination, filters)
	if err != nil {
		logger.L().Error(
			"failed to list products",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse("service_error", "failed to list products"))
		return
	}

	logger.L().Info(
		"products listed successfully",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("total", total),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, response.NewPaginatedResponse(products, total, page, limit, c.Request.URL.Path))
}

// GetProduct handles GET /products/:id requests.
func (h *ProductHandler) GetProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn(
			"ID de produto inválido",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "ID do produto inválido"))
		return
	}

	product, err := h.service.GetProductByID(id)
	if err != nil {
		code, status := CategoryHTTPCode(err)
		logger.L().Error(
			"failed to get product by ID",
			zap.Error(err),
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse(code, err.Error()))
		return
	}

	logger.L().Info(
		"product retrieved by ID",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, product)
}

// GetProductBySlug handles GET /products/slug/:slug requests.
func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	slug := c.Param("slug")

	if slug == "" {
		logger.L().Warn("empty slug in GetProductBySlug", zap.String("request_id", reqID))
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_slug", "product slug is required"))
		return
	}

	product, err := h.service.GetProductBySlug(slug)
	if err != nil {
		code, status := CategoryHTTPCode(err)
		logger.L().Error(
			"failed to get product by slug",
			zap.Error(err),
			zap.String("slug", slug),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse(code, err.Error()))
		return
	}

	logger.L().Info(
		"product retrieved by slug",
		zap.String("slug", slug),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, product)
}

// CreateProduct handles POST /products requests.
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")

	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn(
			"invalid request in CreateProduct",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	product := &domain.Product{
		CategoryID: req.CategoryID,
		Name:       strings.TrimSpace(req.Name),
		Slug:       strings.TrimSpace(req.Slug),
		PriceCents: req.PriceCents,
		IsActive:   req.IsActive,
	}

	if err := h.service.CreateProduct(product); err != nil {
		code, status := CategoryHTTPCode(err)
		logger.L().Error(
			"failed to create product",
			zap.Error(err),
			zap.String("slug", req.Slug),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse(code, err.Error()))
		return
	}

	logger.L().Info(
		"product created successfully",
		zap.Int("product_id", product.ID),
		zap.String("slug", product.Slug),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusCreated, product)
}

// UpdateProduct handles PUT /products/:id requests.
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn(
			"invalid ID in UpdateProduct",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "invalid product ID"))
		return
	}

	var req updateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn(
			"invalid request in UpdateProduct",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	existing, err := h.service.GetProductByID(id)
	if err != nil {
		code, status := CategoryHTTPCode(err)
		logger.L().Error(
			"failed to find product in UpdateProduct",
			zap.Error(err),
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse(code, err.Error()))
		return
	}

	if req.CategoryID != nil {
		existing.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.Slug != nil {
		existing.Slug = strings.TrimSpace(*req.Slug)
	}
	if req.PriceCents != nil {
		existing.PriceCents = *req.PriceCents
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.service.UpdateProduct(id, existing); err != nil {
		code, status := CategoryHTTPCode(err)
		logger.L().Error(
			"failed to update product",
			zap.Error(err),
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse(code, err.Error()))
		return
	}

	updatedProduct, _ := h.service.GetProductByID(id)
	logger.L().Info(
		"product updated successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, updatedProduct)
}

// DeleteProduct handles DELETE /products/:id requests.
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn(
			"invalid ID in DeleteProduct",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "invalid product ID"))
		return
	}

	if err := h.service.DeleteProduct(id); err != nil {
		code, status := CategoryHTTPCode(err)
		logger.L().Error(
			"failed to delete product",
			zap.Error(err),
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse(code, err.Error()))
		return
	}

	logger.L().Info(
		"product deleted successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	c.Status(http.StatusNoContent)
}

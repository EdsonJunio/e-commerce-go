package http

import (
	"net/http"
	"strconv"
	"strings"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/response"
	"e-commerce-go/internal/shared/transport"
	"e-commerce-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service domain.ProductService
}

type createProductRequest struct {
	CategoryID     *int   `json:"category_id" binding:"required,min=1"`
	Name           string `json:"name" binding:"required,min=3,max=255"`
	Slug           string `json:"slug" binding:"required"`
	IsActive       bool   `json:"is_active" binding:"required"`
	SeoTitle       string `json:"seoTitle" binding:"required,min=3,max=255"`
	SeoDescription string `json:"seoDescription" binding:"required,min=3,max=255"`
	Description    string `json:"description" binding:"required,min=3,max=255"`
}

type updateProductRequest struct {
	CategoryID     *int    `json:"category_id,omitempty"`
	Name           *string `json:"name,omitempty"`
	Slug           *string `json:"slug,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
	SeoTitle       *string `json:"seoTitle,omitempty"`
	SeoDescription *string `json:"seoDescription,omitempty"`
	Description    *string `json:"description,omitempty"`
}

func NewProductHandler(service domain.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

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

	products, total, err := h.service.ListProducts(c.Request.Context(), pagination, filters)
	if err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to list products",
			err,
			zap.Int("page", page),
			zap.Int("limit", limit),
			zap.String("request_id", reqID),
		)

		msg := err.Error()
		if mapping.HTTPCode == http.StatusInternalServerError {
			msg = "internal server error"
		}

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	logger.L().Info("products listed successfully",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("total", total),
		zap.String("request_id", reqID))
	c.JSON(http.StatusOK, response.NewPaginatedResponse(products, total, page, limit, c.Request.URL.Path))
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn("invalid ID in GetProduct",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID))
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "invalid product ID"))
		return
	}

	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(mapping, "failed to get product by ID", err,
			zap.Int("id", id),
			zap.String("request_id", reqID))

		msg := err.Error()
		if mapping.HTTPCode == http.StatusInternalServerError {
			msg = "internal server error"
		}

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	logger.L().Info("product retrieved by ID",
		zap.Int("id", id),
		zap.String("request_id", reqID))
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	slug := c.Param("slug")

	if slug == "" {
		logger.L().Warn("empty slug in GetProductBySlug",
			zap.String("request_id", reqID))
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_slug", "product slug is required"))
		return
	}

	product, err := h.service.GetProductBySlug(c.Request.Context(), slug)
	if err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(mapping, "failed to get product by slug",
			err,
			zap.String("slug", slug),
			zap.String("request_id", reqID))

		msg := err.Error()
		if mapping.HTTPCode == http.StatusInternalServerError {
			msg = "internal server error"
		}

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	logger.L().Info("product retrieved by slug",
		zap.String("slug", slug),
		zap.String("request_id", reqID))
	c.JSON(http.StatusOK, product)
}

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
		CategoryID:     req.CategoryID,
		Name:           strings.TrimSpace(req.Name),
		Slug:           strings.TrimSpace(req.Slug),
		IsActive:       req.IsActive,
		SeoTitle:       strings.TrimSpace(req.SeoTitle),
		SeoDescription: strings.TrimSpace(req.SeoDescription),
		Description:    strings.TrimSpace(req.Description),
	}

	if err := h.service.CreateProduct(c.Request.Context(), product); err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to create product",
			err,
			zap.String("slug", req.Slug),
			zap.String("request_id", reqID),
		)

		msg := err.Error()
		if mapping.HTTPCode == http.StatusInternalServerError {
			msg = "internal server error"
		}

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
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

	changes := &domain.Product{}
	if req.CategoryID != nil {
		changes.CategoryID = req.CategoryID
	}
	if req.Name != nil {
		changes.Name = *req.Name
	}
	if req.Slug != nil {
		changes.Slug = *req.Slug
	}
	if req.Description != nil {
		changes.Description = *req.Description
	}
	if req.SeoTitle != nil {
		changes.SeoTitle = *req.SeoTitle
	}
	if req.SeoDescription != nil {
		changes.SeoDescription = *req.SeoDescription
	}
	if req.IsActive != nil {
		changes.IsActive = *req.IsActive
	}

	if err := h.service.UpdateProduct(c.Request.Context(), id, changes); err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to update product",
			err,
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)

		msg := err.Error()
		if mapping.HTTPCode == http.StatusInternalServerError {
			msg = "internal server error"
		}

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	updatedProduct, _ := h.service.GetProductByID(c.Request.Context(), id)

	logger.L().Info(
		"product updated successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, updatedProduct)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn("invalid ID in DeleteProduct",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID))
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "invalid product ID"))
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to delete product",
			err,
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)

		msg := err.Error()
		if mapping.HTTPCode == http.StatusInternalServerError {
			msg = "internal server error"
		}

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	logger.L().Info("product deleted successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID))
	c.Status(http.StatusNoContent)
}

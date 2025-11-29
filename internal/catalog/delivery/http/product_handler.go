package http

import (
	"e-commerce-go/internal/shared/middleware"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/response"
	"e-commerce-go/internal/shared/transport"
	"e-commerce-go/pkg/logger"
)

type ProductHandler struct {
	service domain.ProductService
}

type CreateProductRequest struct {
	CategoryID     *int   `json:"category_id" binding:"required,min=1" example:"1"`
	Name           string `json:"name" binding:"required,min=3,max=255" example:"iPhone 15 Pro"`
	Slug           string `json:"slug" binding:"required" example:"iphone-15-pro"`
	IsActive       bool   `json:"is_active" binding:"required" example:"true"`
	SeoTitle       string `json:"seoTitle" binding:"required,min=3,max=255" example:"iPhone 15 Pro - Apple"`
	SeoDescription string `json:"seoDescription" binding:"required,min=3,max=255" example:"The new iPhone 15 Pro"`
	Description    string `json:"description" binding:"required,min=3,max=255" example:"Titanium design, A17 Pro chip"`
}

type UpdateProductRequest struct {
	CategoryID     *int    `json:"category_id,omitempty" example:"2"`
	Name           *string `json:"name,omitempty" example:"iPhone 15 Pro Max"`
	Slug           *string `json:"slug,omitempty" example:"iphone-15-pro-max"`
	IsActive       *bool   `json:"is_active,omitempty" example:"false"`
	SeoTitle       *string `json:"seoTitle,omitempty" example:"New SEO Title"`
	SeoDescription *string `json:"seoDescription,omitempty" example:"New SEO Desc"`
	Description    *string `json:"description,omitempty" example:"Updated description"`
}

func NewProductHandler(service domain.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) RegisterProductRoutes(router *gin.Engine, auth *middleware.AuthMiddleware) {
	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.GET("", h.ListProducts)
			products.GET("/:id", h.GetProduct)
			products.GET("/slug/:slug", h.GetProductBySlug)

			protected := products.Group("")
			protected.Use(auth.Handle())
			{
				protected.POST("", h.CreateProduct)
				protected.PUT("/:id", h.UpdateProduct)
				protected.DELETE("/:id", h.DeleteProduct)
			}
		}
	}
}

// ListProducts godoc
// @Summary      List all products
// @Description  Get a paginated list of products with optional filters
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        page         query     int     false  "Page number" default(1)
// @Param        limit        query     int     false  "Page size" default(10)
// @Param        category_id  query     int     false  "Filter by Category ID"
// @Param        is_active    query     bool    false  "Filter by Status"
// @Success      200          {object}  response.PaginatedResponse
// @Failure      500          {object}  response.ErrorResponse
// @Router       /products [get]
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

		response.Error(c, mapping.HTTPCode, mapping.Code, msg)
		return
	}

	logger.L().Info("products listed successfully",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("total", total),
		zap.String("request_id", reqID))

	response.SuccessPaginated(c, http.StatusOK, products, total, page, limit)
}

// GetProduct godoc
// @Summary      Get product by ID
// @Description  Retrieve specific product details by its ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  domain.Product
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn("invalid ID in GetProduct",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID))
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product ID")
		return
	}

	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to get product by ID",
			err,
			zap.Int("id", id),
			zap.String("request_id", reqID))

		msg := err.Error()
		if mapping.HTTPCode == http.StatusInternalServerError {
			msg = "internal server error"
		}

		response.Error(c, mapping.HTTPCode, mapping.Code, msg)
		return
	}

	logger.L().Info("product retrieved by ID",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	response.Success(c, http.StatusOK, product)
}

// GetProductBySlug godoc
// @Summary      Get product by Slug
// @Description  Retrieve specific product details by its Slug
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        slug path      string  true  "Product Slug"
// @Success      200  {object}  domain.Product
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /products/slug/{slug} [get]
func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	slug := c.Param("slug")

	if slug == "" {
		logger.L().Warn(
			"empty slug in GetProductBySlug",
			zap.String("request_id", reqID),
		)
		response.Error(c, http.StatusBadRequest, "invalid_slug", "product slug is required")
		return
	}

	product, err := h.service.GetProductBySlug(c.Request.Context(), slug)
	if err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(mapping, "failed to get product by slug",
			err,
			zap.String("slug", slug),
			zap.String("request_id", reqID),
		)

		msg := err.Error()
		if mapping.HTTPCode == http.StatusInternalServerError {
			msg = "internal server error"
		}

		response.Error(c, mapping.HTTPCode, mapping.Code, msg)
		return
	}

	logger.L().Info("product retrieved by slug",
		zap.String("slug", slug),
		zap.String("request_id", reqID))

	response.Success(c, http.StatusOK, product)
}

// CreateProduct godoc
// @Summary      Create a new product
// @Description  Create a new product with the provided data
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        request body      CreateProductRequest  true  "Product Data"
// @Success      201     {object}  domain.Product
// @Failure      400     {object}  response.ErrorResponse
// @Failure      500     {object}  response.ErrorResponse
// @Failure      401     {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn(
			"invalid request in CreateProduct",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
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

		response.Error(c, mapping.HTTPCode, mapping.Code, msg)
		return
	}

	logger.L().Info(
		"product created successfully",
		zap.Int("product_id", product.ID),
		zap.String("slug", product.Slug),
		zap.String("request_id", reqID),
	)
	response.Success(c, http.StatusCreated, product)
}

// UpdateProduct godoc
// @Summary      Update a product
// @Description  Update specific fields of a product by ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      int                   true  "Product ID"
// @Param        request  body      UpdateProductRequest  true  "Update Data"
// @Success      200      {object}  domain.Product
// @Failure      400      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Failure      401     {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /products/{id} [put]
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
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product ID")
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn(
			"invalid request in UpdateProduct",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
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

		response.Error(c, mapping.HTTPCode, mapping.Code, msg)
		return
	}

	updatedProduct, _ := h.service.GetProductByID(c.Request.Context(), id)

	logger.L().Info(
		"product updated successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	response.Success(c, http.StatusOK, updatedProduct)
}

// DeleteProduct godoc
// @Summary      Delete a product
// @Description  Soft delete a product by ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Failure      401     {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn("invalid ID in DeleteProduct",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID))
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product ID")
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

		response.Error(c, mapping.HTTPCode, mapping.Code, msg)
		return
	}

	logger.L().Info(
		"product deleted successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)

	c.Status(http.StatusNoContent)
}

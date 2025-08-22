package http

import (
	"net/http"
	"strconv"
	"strings"

	"e-commerce-go/internal/product/domain"
	"e-commerce-go/internal/shared/response"
	"e-commerce-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service domain.Service
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

func NewProductHandler(service domain.Service) *ProductHandler {
	return &ProductHandler{service: service}
}

// RegisterRoutes registra as rotas do produto no roteador do Gin
func (h *ProductHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("", h.CreateProduct)
			products.GET("", h.ListProducts)
			products.GET("/:id", h.GetProduct)
			products.PUT("/:id", h.UpdateProduct)
			products.DELETE("/:id", h.DeleteProduct)
			products.GET("/slug/:slug", h.GetProductBySlug)
		}
	}
}

// CreateProduct cria um novo produto
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")

	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn("Requisição inválida em CreateProduct",
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
		status := http.StatusInternalServerError
		switch {
		case err.Error() == "product with this slug already exists":
			status = http.StatusConflict
		case strings.Contains(err.Error(), "is required"),
			strings.Contains(err.Error(), "must be greater than zero"):
			status = http.StatusBadRequest
		}

		logger.L().Error("Erro ao criar produto",
			zap.Error(err),
			zap.String("slug", req.Slug),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse("service_error", err.Error()))
		return
	}

	logger.L().Info("Produto criado com sucesso",
		zap.Int("product_id", product.ID),
		zap.String("slug", product.Slug),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusCreated, product)
}

// GetProduct obtém um produto por ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn("ID inválido em GetProduct",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "ID do produto inválido"))
		return
	}

	product, err := h.service.GetProductByID(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "product not found" {
			status = http.StatusNotFound
		}
		logger.L().Error("Erro ao buscar produto por ID",
			zap.Error(err),
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse("service_error", err.Error()))
		return
	}

	logger.L().Info("Produto recuperado por ID",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, product)
}

// GetProductBySlug obtém um produto por slug
func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	slug := c.Param("slug")

	if slug == "" {
		logger.L().Warn("Slug vazio em GetProductBySlug", zap.String("request_id", reqID))
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_slug", "Slug do produto é obrigatório"))
		return
	}

	product, err := h.service.GetProductBySlug(slug)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "product not found" {
			status = http.StatusNotFound
		}
		logger.L().Error("Erro ao buscar produto por slug",
			zap.Error(err),
			zap.String("slug", slug),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse("service_error", err.Error()))
		return
	}

	logger.L().Info("Produto recuperado por slug",
		zap.String("slug", slug),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, product)
}

// UpdateProduct atualiza um produto existente
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn("ID inválido em UpdateProduct",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "ID do produto inválido"))
		return
	}

	var req updateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn("Requisição inválida em UpdateProduct",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	existing, err := h.service.GetProductByID(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "product not found" {
			status = http.StatusNotFound
		}
		logger.L().Error("Erro ao buscar produto existente em UpdateProduct",
			zap.Error(err),
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse("service_error", err.Error()))
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
		status := http.StatusInternalServerError
		switch err.Error() {
		case "product not found":
			status = http.StatusNotFound
		case "product with this slug already exists":
			status = http.StatusConflict
		case "name is required", "slug is required", "price must be greater than zero":
			status = http.StatusBadRequest
		}
		logger.L().Error("Erro ao atualizar produto",
			zap.Error(err),
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse("service_error", err.Error()))
		return
	}

	updatedProduct, _ := h.service.GetProductByID(id)
	logger.L().Info("Produto atualizado com sucesso",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, updatedProduct)
}

// DeleteProduct remove um produto
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn("ID inválido em DeleteProduct",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "ID do produto inválido"))
		return
	}

	if err := h.service.DeleteProduct(id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "product not found" {
			status = http.StatusNotFound
		}
		logger.L().Error("Erro ao deletar produto",
			zap.Error(err),
			zap.Int("id", id),
			zap.String("request_id", reqID),
		)
		c.JSON(status, response.NewErrorResponse("service_error", err.Error()))
		return
	}

	logger.L().Info("Produto excluído com sucesso",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	c.Status(http.StatusNoContent)
}

// ListProducts lista produtos com paginação e filtros
func (h *ProductHandler) ListProducts(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

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

	products, total, err := h.service.ListProducts(limit, page, filters)
	if err != nil {
		logger.L().Error("Erro ao listar produtos",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse("service_error", "Erro ao listar produtos"))
		return
	}

	logger.L().Info("Listagem de produtos realizada",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("total", total),
		zap.String("request_id", reqID),
	)

	c.JSON(http.StatusOK, response.NewPaginatedResponse(
		products,
		total,
		page,
		limit,
		c.Request.URL.Path,
	))
}

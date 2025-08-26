package http

import (
	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/response"
	"e-commerce-go/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CategoryHandler struct {
	service domain.CategoryService
}

type createCategoryRequest struct {
	Name     string `json:"name" binding:"required,min=3,max=255"`
	Slug     string `json:"slug" binding:"required,min=3,max=255"`
	ParentID *uint  `json:"parent_id" binding:"omitempty,gt=0"`
	IsActive bool   `json:"is_active"`
}

type updateCategoryRequest struct {
	Name     *string `json:"name,omitempty"`
	Slug     *string `json:"slug,omitempty"`
	ParentID *uint   `json:"parent_id"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// NewCategoryHandler returns a new CategoryHandler with the given service.
func NewCategoryHandler(service domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// RegisterCategoryRoutes registers category endpoints in the Gin router.
func (h *CategoryHandler) RegisterCategoryRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		category := v1.Group("/categories")
		{
			category.GET("", h.ListCategories)
			category.GET("/:id", h.GetCategory)
			category.GET("/slug/:slug", h.GetCategoryBySlug)
			category.POST("", h.CreateCategory)
			category.PUT("/:id", h.UpdateCategory)
			category.DELETE("/:id", h.DeleteCategory)
		}
	}
}

// ListCategories handles GET /categories with pagination and filters.
func (h *CategoryHandler) ListCategories(c *gin.Context) {
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
	if parentID := c.Query("parent_id"); parentID != "" {
		if parID, err := strconv.Atoi(parentID); err == nil {
			filters["parent_id = ?"] = parID
		}
	}
	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active = ?"] = active
		}
	}

	category, total, err := h.service.ListCategories(limit, page, filters)
	if err != nil {
		logger.L().Error(
			"failed to list categories",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse("service_error", "failed to list categories"))
		return
	}

	logger.L().Info(
		"categories listed successfully",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("total", total),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, response.NewPaginatedResponse(category, total, page, limit, c.Request.URL.Path))

}

// GetCategory handles GET /categories/:id request.
func (h *CategoryHandler) GetCategory(c *gin.Context) {

}

// GetCategoryBySlug handles GET /categories/slug/:slug request.
func (h *CategoryHandler) GetCategoryBySlug(c *gin.Context) {

}

// CreateCategory handles POST /categories request
func (h *CategoryHandler) CreateCategory(c *gin.Context) {

}

// UpdateCategory handler PUT /categories/:id request.
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {

}

// DeleteCategory handles DELETE /categories/:id requests.
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {

}

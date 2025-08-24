package http

import (
	"e-commerce-go/internal/catalog/domain"
	"strconv"

	"github.com/gin-gonic/gin"
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
func (cate *CategoryHandler) RegisterCategoryRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		category := v1.Group("/categories")
		{
			category.GET("", cate.ListCategories)
			category.GET("/:id", cate.GetCategory)
			category.GET("/slug/:slug", cate.GetCategoryBySlug)
			category.POST("", cate.CreateCategory)
			category.PUT("/:id", cate.UpdateCategory)
			category.DELETE("/:id", cate.DeleteCategory)
		}
	}
}

// ListCategories handles GET /categories with pagination and filters.
func (cate *CategoryHandler) ListCategories(c *gin.Context) {
	//reqID := c.Writer.Header().Get("X-Request-ID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	//	filter := make(map[string]interface{})
	//	if cate
}

// GetCategory handles GET /categories/:id request.
func (cate *CategoryHandler) GetCategory(c *gin.Context) {

}

// GetCategoryBySlug handles GET /categories/slug/:slug request.
func (cate *CategoryHandler) GetCategoryBySlug(c *gin.Context) {

}

// CreateCategory handles POST /categories request
func (cate *CategoryHandler) CreateCategory(c *gin.Context) {

}

// UpdateCategory handler PUT /categories/:id request.
func (cate *CategoryHandler) UpdateCategory(c *gin.Context) {

}

// DeleteCategory handles DELETE /categories/:id requests.
func (cate *CategoryHandler) DeleteCategory(c *gin.Context) {

}

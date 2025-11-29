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

type CategoryHandler struct {
	service domain.CategoryService
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=255" example:"Electronics"`
	Slug        string `json:"slug" binding:"required" example:"electronics"`
	ParentID    *int   `json:"parent_id" example:"1"`
	IsActive    bool   `json:"is_active" example:"true"`
	Description string `json:"description" example:"All kinds of electronic devices"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty" example:"Home Appliances"`
	Slug        *string `json:"slug,omitempty" example:"home-appliances"`
	ParentID    *uint   `json:"parent_id,omitempty" example:"2"`
	IsActive    *bool   `json:"is_active,omitempty" example:"false"`
	Description *string `json:"description,omitempty" example:"Updated description"`
}

func NewCategoryHandler(service domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) RegisterCategoryRoutes(router *gin.Engine, auth *middleware.AuthMiddleware) {
	v1 := router.Group("/api/v1")
	{
		category := v1.Group("/categories")
		{
			category.GET("", h.ListCategories)
			category.GET("/:id", h.GetCategory)
			category.GET("/slug/:slug", h.GetCategoryBySlug)

			protected := category.Group("")
			protected.Use(auth.Handle())
			{
				protected.POST("", h.CreateCategory)
				protected.PUT("/:id", h.UpdateCategory)
				protected.DELETE("/:id", h.DeleteCategory)
			}
		}
	}
}

// ListCategories godoc
// @Summary      List all categories
// @Description  Get a paginated list of categories with optional filters
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "Page number" default(1)
// @Param        limit      query     int     false  "Page size" default(10)
// @Param        parent_id  query     int     false  "Filter by Parent ID"
// @Param        is_active  query     bool    false  "Filter by Status"
// @Success      200        {object}  response.PaginatedResponse
// @Failure      500        {object}  response.ErrorResponse
// @Router       /categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	pagination := domain.NewPagination(page, limit)

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

	categories, total, err := h.service.ListCategories(c.Request.Context(), pagination, filters)
	if err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to list categories",
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

	logger.L().Info("categories listed successfully",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("total", total),
		zap.String("request_id", reqID))

	response.SuccessPaginated(c, http.StatusOK, categories, total, page, limit)
}

// GetCategory godoc
// @Summary      Get category by ID
// @Description  Retrieve specific category details by its ID
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  domain.Category
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /categories/{id} [get]
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn(
			"invalid category ID",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		response.Error(c, http.StatusBadRequest, "invalid_id", "category ID is invalid")
		return
	}

	category, err := h.service.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to get category by ID",
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
		"category retrieved by ID",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	response.Success(c, http.StatusOK, category)
}

// GetCategoryBySlug godoc
// @Summary      Get category by Slug
// @Description  Retrieve specific category details by its Slug
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        slug path      string  true  "Category Slug"
// @Success      200  {object}  domain.Category
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /categories/slug/{slug} [get]
func (h *CategoryHandler) GetCategoryBySlug(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	slug := c.Param("slug")

	if slug == "" {
		logger.L().Warn(
			"empty slug in GetCategoryBySlug",
			zap.String("request_id", reqID),
		)
		response.Error(c, http.StatusBadRequest, "invalid_slug", "category slug is required")
		return
	}

	categorySlug, err := h.service.GetCategoryBySlug(c.Request.Context(), slug)
	if err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(mapping,
			"failed to get category",
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

	response.Success(c, http.StatusOK, categorySlug)
}

// CreateCategory godoc
// @Summary      Create a new category
// @Description  Create a new category with the provided data
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        request body      CreateCategoryRequest  true  "Category Data"
// @Success      201     {object}  domain.Category
// @Failure      400     {object}  response.ErrorResponse
// @Failure      401     {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn("invalid request in CreateCategory", zap.Error(err), zap.String("request_id", reqID))
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	category := &domain.Category{
		Name:        strings.TrimSpace(req.Name),
		Slug:        strings.TrimSpace(req.Slug),
		IsActive:    req.IsActive,
		ParentID:    req.ParentID,
		Description: strings.TrimSpace(req.Description),
	}

	if err := h.service.CreateCategory(c.Request.Context(), category); err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to create category",
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

	response.Success(c, http.StatusCreated, category)
}

// UpdateCategory godoc
// @Summary      Update a category
// @Description  Update specific fields of a category by ID
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Category ID"
// @Param        request  body      UpdateCategoryRequest  true  "Update Data"
// @Success      200      {object}  domain.Category
// @Failure      400      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Failure      401     {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn(
			"invalid category ID in updateCategory",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		response.Error(c, http.StatusBadRequest, "invalid_id", "category ID must be a valid integer")
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn(
			"Invalid request body in updateCategory",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	updateData := &domain.Category{
		ID: id,
	}

	if req.Name != nil {
		updateData.Name = strings.TrimSpace(*req.Name)
	}
	if req.Slug != nil {
		updateData.Slug = strings.TrimSpace(*req.Slug)
	}
	if req.ParentID != nil {
		parentIDValue := int(*req.ParentID)
		updateData.ParentID = &parentIDValue
	}
	if req.IsActive != nil {
		updateData.IsActive = *req.IsActive
	}
	if req.Description != nil {
		updateData.Description = *req.Description
	}

	if err := h.service.UpdateCategory(c.Request.Context(), id, updateData); err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to update category",
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

	updatedCategory, _ := h.service.GetCategoryByID(c.Request.Context(), id)

	logger.L().Info(
		"category updated successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)

	response.Success(c, http.StatusOK, updatedCategory)
}

// DeleteCategory godoc
// @Summary      Delete a category
// @Description  Soft delete a category by ID
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Failure      401     {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn(
			"invalid category ID in DeleteCategory",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		response.Error(c, http.StatusBadRequest, "invalid_id", "category ID must be a valid integer")
		return
	}

	if err := h.service.DeleteCategory(c.Request.Context(), id); err != nil {
		mapping := transport.HTTPErrorMapper(err)

		transport.LogByErrorMapping(
			mapping,
			"failed to delete category",
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
		"category deleted successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)

	c.Status(http.StatusNoContent)
}

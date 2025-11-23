package http

import (
	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/response"
	"e-commerce-go/internal/shared/transport"
	"e-commerce-go/pkg/logger"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CategoryHandler struct {
	service domain.CategoryService
}

type createCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=255"`
	Slug        string `json:"slug" binding:"required"`
	ParentID    *int   `json:"parent_id"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
}

type updateCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	ParentID    *uint   `json:"parent_id,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	Description *string `json:"description,omitempty"`
}

func NewCategoryHandler(service domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

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

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	logger.L().Info("products listed successfully",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("total", total),
		zap.String("request_id", reqID))
	c.JSON(http.StatusOK, response.NewPaginatedResponse(categories, total, page, limit, c.Request.URL.Path))
}

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
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "category ID is invalid"))
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

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	logger.L().Info(
		"category retrieved by ID",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)
	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) GetCategoryBySlug(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")
	slug := c.Param("slug")

	if slug == "" {
		logger.L().Warn(
			"empty slug in GetCategoryBySlug",
			zap.String("request_id", reqID),
		)

		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_slug", "category slug is required"))
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

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	c.JSON(http.StatusOK, categorySlug)
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")

	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn("invalid request in CreateCategory", zap.Error(err), zap.String("request_id", reqID))
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_request", err.Error()))
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

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	reqID := c.Writer.Header().Get("X-Request-ID")

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.L().Warn(
			"invalid category ID in UpdateCategory",
			zap.String("id_param", idParam),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "category ID must be a valid integer"))
		return
	}

	var req updateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Warn(
			"invalid request body in updateCategory",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_request", err.Error()))
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

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	updatedCategory, _ := h.service.GetCategoryByID(c.Request.Context(), id)

	logger.L().Info(
		"category updated successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)

	c.JSON(http.StatusOK, updatedCategory)
}

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
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("invalid_id", "category ID must be a valid integer"))
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

		c.JSON(mapping.HTTPCode, response.NewErrorResponse(mapping.Code, msg))
		return
	}

	logger.L().Info(
		"category deleted successfully",
		zap.Int("id", id),
		zap.String("request_id", reqID),
	)

	c.Status(http.StatusNoContent)
}

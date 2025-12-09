package http_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	delivery "e-commerce-go/internal/catalog/delivery/http"
	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/catalog/mocks"
	"e-commerce-go/pkg/logger"
)

func TestProductHandler_ListProducts(t *testing.T) {
	_ = logger.Init(logger.Config{
		Environment: "test",
		Service:     "test-service",
		Version:     "1.0",
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockProductService(ctrl)
	handler := delivery.NewProductHandler(mockService)
	gin.SetMode(gin.TestMode)

	dummyProducts := []domain.Product{
		{ID: 1, Name: "Product A", Slug: "prod-a"},
		{ID: 2, Name: "Product B", Slug: "prod-b"},
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - Default Pagination (No Params)",
			queryParams: "",
			setupMock: func() {
				expectedPagination := domain.NewPagination(1, 10)
				expectedFilters := map[string]interface{}{}

				mockService.EXPECT().
					ListProducts(gomock.Any(), expectedPagination, expectedFilters).
					Return(dummyProducts, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"total":2`,
		},
		{
			name:        "Success - With Filters and Custom Pagination",
			queryParams: "?page=2&limit=5&category_id=10&is_active=true",
			setupMock: func() {
				expectedPagination := domain.NewPagination(2, 5)
				expectedFilters := map[string]interface{}{
					"category_id = ?": 10,
					"is_active = ?":   true,
				}

				mockService.EXPECT().
					ListProducts(gomock.Any(), expectedPagination, expectedFilters).
					Return([]domain.Product{}, int64(0), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"page":2`,
		},
		{
			name:        "Error - Service Fail",
			queryParams: "",
			setupMock: func() {
				mockService.EXPECT().
					ListProducts(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, int64(0), errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `internal server error`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/products"+tt.queryParams, nil)

			handler.ListProducts(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

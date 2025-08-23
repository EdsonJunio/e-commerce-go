package http_test

import (
	"bytes"
	producthttp "e-commerce-go/internal/catalog/delivery/http"
	"encoding/json"

	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/response"
	"e-commerce-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	// Initialize test logger
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	testLogger, _ := zapCfg.Build()

	// Override the global logger with test configuration
	logger.Init(logger.Config{
		Environment: "test",
		Service:     "product-handler-test",
		Version:     "test",
	})

	// Replace the global logger with our test logger
	zap.ReplaceGlobals(testLogger)
}

type mockService struct {
	mock.Mock
}

func (m *mockService) CreateProduct(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockService) GetProductByID(id int) (*domain.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockService) GetProductBySlug(slug string) (*domain.Product, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockService) UpdateProduct(id int, product *domain.Product) error {
	args := m.Called(id, product)
	return args.Error(0)
}

func (m *mockService) DeleteProduct(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockService) ListProducts(limit, page int, filters map[string]interface{}) ([]domain.Product, int64, error) {
	args := m.Called(limit, page, filters)
	return args.Get(0).([]domain.Product), args.Get(1).(int64), args.Error(2)
}

func setupRouter(service domain.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := producthttp.NewProductHandler(service)
	handler.RegisterRoutes(r)
	return r
}

func TestProductHandler_CreateProduct(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name         string
		requestBody  interface{}
		setupMock    func(*mockService)
		expectedCode int
		expectedBody interface{}
		isError      bool
	}{
		{
			name: "successful creation",
			requestBody: gin.H{
				"name":        "Test Product",
				"slug":        "test-product",
				"price_cents": 1000,
				"category_id": 1,
				"is_active":   true,
			},
			setupMock: func(ms *mockService) {
				ms.On("CreateProduct", mock.AnythingOfType("*domain.Product")).
					Run(func(args mock.Arguments) {
						p := args.Get(0).(*domain.Product)
						p.ID = 1
						p.CreatedAt = now
						p.UpdatedAt = now
					}).
					Return(nil)
			},
			expectedCode: http.StatusCreated,
			expectedBody: gin.H{
				"id":          float64(1),
				"name":        "Test Product",
				"slug":        "test-product",
				"price_cents": float64(1000),
				"category_id": float64(1),
				"is_active":   true,
			},
			isError: false,
		},
		{
			name: "invalid request body",
			requestBody: gin.H{
				"name": "",
			},
			setupMock:    func(ms *mockService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: response.ErrorResponse{
				Code:    "invalid_request",
				Message: "Key: 'createProductRequest.CategoryID' Error:Field validation for 'CategoryID' failed on the 'required' tag\nKey: 'createProductRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag\nKey: 'createProductRequest.PriceCents' Error:Field validation for 'PriceCents' failed on the 'required' tag\nKey: 'createProductRequest.Slug' Error:Field validation for 'Slug' failed on the 'required' tag",
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockService)
			tt.setupMock(ms)

			r := setupRouter(ms)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			var responseBody map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &responseBody)

			if tt.isError {
				expected := tt.expectedBody.(response.ErrorResponse)
				assert.Equal(t, expected.Code, responseBody["code"])
				message := responseBody["message"].(string)
				assert.Contains(t, message, "CategoryID")
				assert.Contains(t, message, "Name")
				assert.Contains(t, message, "PriceCents")
				assert.Contains(t, message, "Slug")
			} else {
				expected := make(map[string]interface{})
				switch v := tt.expectedBody.(type) {
				case map[string]interface{}:
					expected = v
				case gin.H:
					for k, v := range v {
						expected[k] = v
					}
				default:
					t.Fatalf("unexpected type: %T", tt.expectedBody)
				}

				for key, expectedValue := range expected {
					assert.Equal(t, expectedValue, responseBody[key], "mismatch in field: %s", key)
				}
			}

			ms.AssertExpectations(t)
		})
	}
}

func TestProductHandler_GetProduct(t *testing.T) {
	tests := []struct {
		name         string
		productID    string
		setupMock    func(*mockService)
		expectedCode int
		expectedBody interface{}
	}{
		{
			name:      "product found",
			productID: "1",
			setupMock: func(ms *mockService) {
				ms.On("GetProductByID", 1).Return(&domain.Product{
					ID:         1,
					Name:       "Test Product",
					Slug:       "test-product",
					PriceCents: 1000,
					CategoryID: 1,
					IsActive:   true,
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: gin.H{
				"id":          float64(1),
				"name":        "Test Product",
				"slug":        "test-product",
				"price_cents": float64(1000),
				"category_id": float64(1),
				"is_active":   true,
			},
		},
		{
			name:      "product not found",
			productID: "999",
			setupMock: func(ms *mockService) {
				ms.On("GetProductByID", 999).Return((*domain.Product)(nil), domain.ErrNotFound)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: response.ErrorResponse{
				Code:    "service_error",
				Message: "product not found",
			},
		},
		{
			name:         "invalid product ID",
			productID:    "invalid",
			setupMock:    func(ms *mockService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: response.ErrorResponse{
				Code:    "invalid_id",
				Message: "ID do produto inválido",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockService)
			tt.setupMock(ms)

			r := setupRouter(ms)

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/products/"+tt.productID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			var responseBody map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &responseBody)

			switch expected := tt.expectedBody.(type) {
			case response.ErrorResponse:
				assert.Equal(t, expected.Code, responseBody["code"])
				assert.Equal(t, expected.Message, responseBody["message"])
			case map[string]interface{}:
				delete(responseBody, "created_at")
				delete(responseBody, "updated_at")
				delete(expected, "created_at")
				delete(expected, "updated_at")
				assert.Equal(t, expected, responseBody)
			}

			ms.AssertExpectations(t)
		})
	}
}

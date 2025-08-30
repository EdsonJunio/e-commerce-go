package service_test

import (
	"e-commerce-go/internal/catalog/service"
	"testing"
	"time"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/pkg/logger"

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
		Service:     "product-service-test",
		Version:     "test",
	})

	// Replace the global logger with our test logger
	zap.ReplaceGlobals(testLogger)
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockRepository) FindByID(id int) (*domain.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockRepository) FindBySlug(slug string) (*domain.Product, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockRepository) Update(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockRepository) List(limit, offset int, filters map[string]interface{}) ([]domain.Product, int64, error) {
	args := m.Called(limit, offset, filters)
	return args.Get(0).([]domain.Product), args.Get(1).(int64), args.Error(2)
}

func TestProductService_CreateProduct(t *testing.T) {
	tests := []struct {
		name        string
		product     *domain.Product
		setupMock   func(*mockRepository)
		expectError bool
		errMsg      string
	}{
		{
			name: "successful creation",
			product: &domain.Product{
				Name:       "Test Product",
				Slug:       "test-product",
				PriceCents: 1000,
				CategoryID: 1,
			},
			setupMock: func(mr *mockRepository) {
				mr.On("FindBySlug", "test-product").Return((*domain.Product)(nil), nil)
				mr.On("Create", mock.AnythingOfType("*domain.Product")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "empty name",
			product: &domain.Product{
				Name:       "",
				Slug:       "test-product",
				PriceCents: 1000,
			},
			setupMock:   func(mr *mockRepository) {},
			expectError: true,
			errMsg:      "product name is required",
		},
		{
			name: "duplicate slug",
			product: &domain.Product{
				Name:       "Test Product",
				Slug:       "existing-product",
				PriceCents: 1000,
			},
			setupMock: func(mr *mockRepository) {
				existing := &domain.Product{ID: 1, Slug: "existing-product"}
				mr.On("FindBySlug", "existing-product").Return(existing, nil)
			},
			expectError: true,
			errMsg:      "product with this slug already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := new(mockRepository)
			tt.setupMock(mr)

			svc := service.NewProductService(mr)
			err := svc.CreateProduct(tt.product)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			mr.AssertExpectations(t)
		})
	}
}

func TestProductService_GetProductByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		setupMock   func(*mockRepository)
		expectError bool
		errMsg      string
	}{
		{
			name: "product found",
			id:   1,
			setupMock: func(mr *mockRepository) {
				mr.On("FindByID", 1).Return(&domain.Product{ID: 1, Name: "Test Product"}, nil)
			},
			expectError: false,
		},
		{
			name: "product not found",
			id:   999,
			setupMock: func(mr *mockRepository) {
				mr.On("FindByID", 999).Return((*domain.Product)(nil), nil)
			},
			expectError: true,
			errMsg:      "product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := new(mockRepository)
			tt.setupMock(mr)

			svc := service.NewProductService(mr)
			product, err := svc.GetProductByID(tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, product)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, tt.id, product.ID)
			}
			mr.AssertExpectations(t)
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name        string
		id          int
		product     *domain.Product
		setupMock   func(*mockRepository)
		expectError bool
		errMsg      string
	}{
		{
			name: "successful update",
			id:   1,
			product: &domain.Product{
				Name:       "Updated Product",
				Slug:       "updated-product",
				PriceCents: 2000,
				CategoryID: 2,
				IsActive:   true,
			},
			setupMock: func(mr *mockRepository) {
				existing := &domain.Product{
					ID:         1,
					Name:       "Old Product",
					Slug:       "old-product",
					PriceCents: 1000,
					CategoryID: 1,
					IsActive:   false,
					CreatedAt:  testTime,
				}
				mr.On("FindByID", 1).Return(existing, nil)
				mr.On("Update", mock.MatchedBy(func(p *domain.Product) bool {
					return p.ID == 1 &&
						p.Name == "Updated Product" &&
						p.Slug == "updated-product" &&
						p.PriceCents == 2000 &&
						p.CategoryID == 2 &&
						p.IsActive == true &&
						p.CreatedAt == testTime
				})).Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := new(mockRepository)
			tt.setupMock(mr)

			svc := service.NewProductService(mr)
			err := svc.UpdateProduct(tt.id, tt.product)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			mr.AssertExpectations(t)
		})
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		setupMock   func(*mockRepository)
		expectError bool
		errMsg      string
	}{
		{
			name: "successful deletion",
			id:   1,
			setupMock: func(mr *mockRepository) {
				mr.On("FindByID", 1).Return(&domain.Product{ID: 1}, nil)
				mr.On("Delete", 1).Return(nil)
			},
			expectError: false,
		},
		{
			name: "product not found",
			id:   999,
			setupMock: func(mr *mockRepository) {
				mr.On("FindByID", 999).Return((*domain.Product)(nil), nil)
			},
			expectError: true,
			errMsg:      "product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := new(mockRepository)
			tt.setupMock(mr)

			svc := service.NewProductService(mr)
			err := svc.DeleteProduct(tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			mr.AssertExpectations(t)
		})
	}
}

func TestProductService_ListProducts(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		page        int
		filters     map[string]interface{}
		setupMock   func(*mockRepository)
		expectError bool
		errMsg      string
	}{
		{
			name:  "successful listing",
			limit: 10,
			page:  1,
			filters: map[string]interface{}{
				"is_active": true,
			},
			setupMock: func(mr *mockRepository) {
				products := []domain.Product{
					{ID: 1, Name: "Product 1"},
					{ID: 2, Name: "Product 2"},
				}
				mr.On("List", 10, 0, map[string]interface{}{"is_active": true}).Return(products, int64(2), nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := new(mockRepository)
			tt.setupMock(mr)

			svc := service.NewProductService(mr)
			pagination := domain.NewPagination(tt.page, tt.limit)
			products, total, err := svc.ListProducts(pagination, tt.filters)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, products)
				assert.Zero(t, total)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, products)
				assert.Greater(t, len(products), 0)
				assert.Greater(t, total, int64(0))
			}
			mr.AssertExpectations(t)
		})
	}
}

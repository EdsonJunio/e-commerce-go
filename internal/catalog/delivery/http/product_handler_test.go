package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type productServiceStub struct {
	listCalls  int
	pagination domain.Pagination
	filters    domain.ProductListFilters
}

func (s *productServiceStub) ListProducts(_ context.Context, p domain.Pagination, filters domain.ProductListFilters) ([]domain.Product, int64, error) {
	s.listCalls++
	s.pagination = p
	s.filters = filters
	return nil, 0, nil
}
func (*productServiceStub) GetProductByID(context.Context, int) (*domain.Product, error) {
	panic("not used")
}
func (*productServiceStub) GetProductBySlug(context.Context, string) (*domain.Product, error) {
	panic("not used")
}
func (*productServiceStub) CreateProduct(context.Context, *domain.Product) error { panic("not used") }
func (*productServiceStub) UpdateProduct(context.Context, int, *domain.Product) error {
	panic("not used")
}
func (*productServiceStub) DeleteProduct(context.Context, int) error { panic("not used") }

func productListRequest(service domain.ProductService, query string) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/api/v1/products", NewProductHandler(service).ListProducts)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/products?"+query, nil))
	return recorder
}

func TestProductListFiltersReachService(t *testing.T) {
	tests := []struct {
		name, query string
		category    *int
		active      *bool
	}{
		{"unfiltered", "page=2&limit=25", nil, nil},
		{"category", "category_id=12", intPointer(12), nil},
		{"active true", "is_active=true", nil, boolPointer(true)},
		{"active false", "is_active=false", nil, boolPointer(false)},
		{"combined", "category_id=12&is_active=false", intPointer(12), boolPointer(false)},
		{"repeated first value", "category_id=12&category_id=24&is_active=false&is_active=true", intPointer(12), boolPointer(false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &productServiceStub{}
			res := productListRequest(service, tt.query)
			if res.Code != http.StatusOK || service.listCalls != 1 {
				t.Fatalf("status=%d calls=%d", res.Code, service.listCalls)
			}
			assertOptionalIntFilter(t, service.filters.CategoryID, tt.category)
			assertOptionalBoolFilter(t, service.filters.IsActive, tt.active)
			if tt.name == "unfiltered" && service.pagination != (domain.Pagination{Page: 2, Limit: 25, Offset: 25}) {
				t.Errorf("pagination=%+v", service.pagination)
			}
		})
	}
}

func TestProductListRejectsMalformedFilters(t *testing.T) {
	for _, query := range []string{"category_id=", "category_id=abc", "category_id=0", "category_id=-1", "category_id=999999999999999999999999999", "is_active=", "is_active=maybe"} {
		t.Run(query, func(t *testing.T) {
			service := &productServiceStub{}
			res := productListRequest(service, query)
			if res.Code != http.StatusBadRequest || service.listCalls != 0 {
				t.Fatalf("status=%d calls=%d", res.Code, service.listCalls)
			}
			var body response.ErrorResponse
			if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != "invalid_request" {
				t.Errorf("code=%q", body.Code)
			}
		})
	}
}

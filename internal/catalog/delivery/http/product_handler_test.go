package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/shared/middleware"
	"e-commerce-go/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type productServiceStub struct {
	listCalls   int
	pagination  domain.Pagination
	filters     domain.ProductListFilters
	updateCalls int
	changes     domain.ProductChanges
	updateErr   error
}

func (s *productServiceStub) ListProducts(_ context.Context, p domain.Pagination, filters domain.ProductListFilters) ([]domain.Product, int64, error) {
	s.listCalls++
	s.pagination = p
	s.filters = filters
	return nil, 0, nil
}
func (*productServiceStub) GetProductByID(context.Context, int) (*domain.Product, error) {
	return &domain.Product{ID: 1}, nil
}
func (*productServiceStub) GetProductBySlug(context.Context, string) (*domain.Product, error) {
	panic("not used")
}
func (*productServiceStub) CreateProduct(context.Context, *domain.Product) error { panic("not used") }
func (s *productServiceStub) UpdateProduct(_ context.Context, _ int, changes domain.ProductChanges) error {
	s.updateCalls++
	s.changes = changes
	return s.updateErr
}
func (*productServiceStub) DeleteProduct(context.Context, int) error { panic("not used") }

func productUpdateRequest(service domain.ProductService, body string) *httptest.ResponseRecorder {
	router := gin.New()
	router.PUT("/api/v1/products/:id", NewProductHandler(service).UpdateProduct)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/products/1", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	result := httptest.NewRecorder()
	router.ServeHTTP(result, request)
	return result
}

func TestProductHandlerUpdatePreservesFieldPresence(t *testing.T) {
	for _, tt := range []struct {
		name  string
		body  string
		check func(*testing.T, domain.ProductChanges)
	}{
		{"omitted fields", `{"name":"After"}`, func(t *testing.T, c domain.ProductChanges) {
			if c.Name == nil || *c.Name != "After" || c.IsActive != nil || c.CategoryID != nil || c.Slug != nil {
				t.Fatalf("changes = %+v", c)
			}
		}},
		{"explicit false", `{"is_active":false}`, func(t *testing.T, c domain.ProductChanges) {
			if c.IsActive == nil || *c.IsActive {
				t.Fatalf("changes = %+v", c)
			}
		}},
		{"category", `{"category_id":3}`, func(t *testing.T, c domain.ProductChanges) {
			if c.CategoryID == nil || *c.CategoryID != 3 {
				t.Fatalf("changes = %+v", c)
			}
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service := &productServiceStub{}
			result := productUpdateRequest(service, tt.body)
			if result.Code != http.StatusOK || service.updateCalls != 1 {
				t.Fatalf("status = %d, calls = %d, body = %s", result.Code, service.updateCalls, result.Body.String())
			}
			tt.check(t, service.changes)
		})
	}
}

func TestProductHandlerUpdateRejectsMalformedAndMissing(t *testing.T) {
	for _, body := range []string{`{"is_active":"false"}`, `{"category_id":"3"}`, `{"name":3}`} {
		t.Run(body, func(t *testing.T) {
			service := &productServiceStub{}
			result := productUpdateRequest(service, body)
			if result.Code != http.StatusBadRequest || service.updateCalls != 0 {
				t.Fatalf("status = %d, calls = %d", result.Code, service.updateCalls)
			}
		})
	}
	service := &productServiceStub{updateErr: domain.ErrProductNotFound}
	result := productUpdateRequest(service, `{}`)
	if result.Code != http.StatusNotFound || service.updateCalls != 1 {
		t.Fatalf("status = %d, calls = %d", result.Code, service.updateCalls)
	}
}

func TestProductUpdateRouteAuthorization(t *testing.T) {
	for _, tt := range []struct {
		name, role, authorization string
		want                      int
	}{
		{"anonymous", "", "", http.StatusUnauthorized},
		{"customer", "customer", "Bearer valid", http.StatusForbidden},
		{"admin", "admin", "Bearer valid", http.StatusOK},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service := &productServiceStub{}
			router := gin.New()
			NewProductHandler(service).RegisterProductRoutes(router, middleware.NewAuthMiddleware(&categoryTokenValidator{role: tt.role}))
			request := httptest.NewRequest(http.MethodPut, "/api/v1/products/1", strings.NewReader(`{"is_active":false}`))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", tt.authorization)
			result := httptest.NewRecorder()
			router.ServeHTTP(result, request)
			if result.Code != tt.want {
				t.Fatalf("status = %d, want %d", result.Code, tt.want)
			}
			wantCalls := 0
			if tt.role == "admin" {
				wantCalls = 1
			}
			if service.updateCalls != wantCalls {
				t.Fatalf("update calls = %d, want %d", service.updateCalls, wantCalls)
			}
		})
	}
}

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

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
	"e-commerce-go/internal/shared/service"
	"e-commerce-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

type categoryServiceStub struct {
	listCalls   int
	pagination  domain.Pagination
	filters     domain.CategoryListFilters
	changes     domain.CategoryChanges
	updateCalls int
	updateErr   error
	getCalls    int
	getErr      error
}

func (s *categoryServiceStub) ListCategories(_ context.Context, pagination domain.Pagination, filters domain.CategoryListFilters) ([]domain.Category, int64, error) {
	s.listCalls++
	s.pagination = pagination
	s.filters = filters
	return []domain.Category{{ID: 1, Name: "Electronics", IsActive: true}}, 1, nil
}

func (s *categoryServiceStub) GetCategoryByID(context.Context, int) (*domain.Category, error) {
	s.getCalls++
	if s.getErr != nil {
		return nil, s.getErr
	}
	return &domain.Category{ID: 1, Name: "Category", Slug: "category", Description: "Description"}, nil
}

func (*categoryServiceStub) GetCategoryBySlug(context.Context, string) (*domain.Category, error) {
	panic("not used")
}

func (*categoryServiceStub) CreateCategory(context.Context, *domain.Category) error {
	panic("not used")
}

func (s *categoryServiceStub) UpdateCategory(_ context.Context, _ int, changes domain.CategoryChanges) error {
	s.updateCalls++
	s.changes = changes
	return s.updateErr
}

func TestCategoryHandlerUpdateRejectsHierarchyCycle(t *testing.T) {
	service := &categoryServiceStub{updateErr: domain.ErrInvalidCategoryReference}
	router := gin.New()
	router.PUT("/api/v1/categories/:id", NewCategoryHandler(service).UpdateCategory)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/categories/1", strings.NewReader(`{"parent_id":3}`))
	request.Header.Set("Content-Type", "application/json")
	result := httptest.NewRecorder()
	router.ServeHTTP(result, request)
	if result.Code != http.StatusBadRequest || service.updateCalls != 1 {
		t.Fatalf("status = %d, calls = %d, body = %s", result.Code, service.updateCalls, result.Body.String())
	}
	if !strings.Contains(result.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("unexpected error code: %s", result.Body.String())
	}
}

func TestCategoryHandlerUpdatePropagatesPostWriteReadFailure(t *testing.T) {
	service := &categoryServiceStub{getErr: domain.ErrCategoryNotFound}
	router := gin.New()
	router.PUT("/api/v1/categories/:id", NewCategoryHandler(service).UpdateCategory)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/categories/1", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	result := httptest.NewRecorder()

	router.ServeHTTP(result, request)

	if result.Code != http.StatusNotFound || service.updateCalls != 1 || service.getCalls != 1 {
		t.Fatalf("status = %d, update calls = %d, get calls = %d, body = %s", result.Code, service.updateCalls, service.getCalls, result.Body.String())
	}
	if !strings.Contains(result.Body.String(), `"code":"not_found"`) {
		t.Fatalf("unexpected error response: %s", result.Body.String())
	}
}

func (*categoryServiceStub) DeleteCategory(context.Context, int) error {
	panic("not used")
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	_ = logger.Init(logger.Config{Environment: "test", Service: "catalog-http-test", Version: "test"})
	m.Run()
}

func TestCategoryHandlerUpdatePresence(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		check func(*testing.T, domain.CategoryChanges)
	}{
		{"omitted fields", `{"name":"New"}`, func(t *testing.T, c domain.CategoryChanges) {
			if c.Name == nil || *c.Name != "New" || c.IsActive != nil || c.ParentID != nil || c.ClearParent {
				t.Fatalf("changes = %+v", c)
			}
		}},
		{"explicit false", `{"is_active":false}`, func(t *testing.T, c domain.CategoryChanges) {
			if c.IsActive == nil || *c.IsActive {
				t.Fatalf("changes = %+v", c)
			}
		}},
		{"clear parent", `{"parent_id":null}`, func(t *testing.T, c domain.CategoryChanges) {
			if !c.ClearParent || c.ParentID != nil {
				t.Fatalf("changes = %+v", c)
			}
		}},
		{"assign parent", `{"parent_id":2}`, func(t *testing.T, c domain.CategoryChanges) {
			if c.ClearParent || c.ParentID == nil || *c.ParentID != 2 {
				t.Fatalf("changes = %+v", c)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &categoryServiceStub{}
			router := gin.New()
			router.PUT("/api/v1/categories/:id", NewCategoryHandler(service).UpdateCategory)
			request := httptest.NewRequest(http.MethodPut, "/api/v1/categories/1", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			result := httptest.NewRecorder()
			router.ServeHTTP(result, request)
			if result.Code != http.StatusOK || service.updateCalls != 1 {
				t.Fatalf("status = %d, calls = %d, body = %s", result.Code, service.updateCalls, result.Body.String())
			}
			tt.check(t, service.changes)
		})
	}
}

func TestCategoryHandlerUpdateRejectsInvalidParent(t *testing.T) {
	for _, body := range []string{`{"parent_id":0}`, `{"parent_id":-1}`, `{"parent_id":"2"}`, `{"parent_id":1.5}`} {
		t.Run(body, func(t *testing.T) {
			service := &categoryServiceStub{}
			router := gin.New()
			router.PUT("/api/v1/categories/:id", NewCategoryHandler(service).UpdateCategory)
			request := httptest.NewRequest(http.MethodPut, "/api/v1/categories/1", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			result := httptest.NewRecorder()
			router.ServeHTTP(result, request)
			if result.Code != http.StatusBadRequest || service.updateCalls != 0 {
				t.Fatalf("status = %d, calls = %d", result.Code, service.updateCalls)
			}
		})
	}
}

type categoryTokenValidator struct{ role string }

func (*categoryTokenValidator) GenerateToken(int, string) (string, error) { return "", nil }
func (v *categoryTokenValidator) ValidateToken(string) (*service.TokenClaims, error) {
	return &service.TokenClaims{UserID: 1, Role: v.role}, nil
}

func TestCategoryUpdateRouteAuthorization(t *testing.T) {
	for _, tt := range []struct {
		name          string
		role          string
		authorization string
		want          int
	}{
		{"anonymous", "", "", http.StatusUnauthorized},
		{"customer", "customer", "Bearer valid", http.StatusForbidden},
		{"admin", "admin", "Bearer valid", http.StatusOK},
	} {
		t.Run(tt.name, func(t *testing.T) {
			catalogService := &categoryServiceStub{}
			router := gin.New()
			NewCategoryHandler(catalogService).RegisterCategoryRoutes(router, middleware.NewAuthMiddleware(&categoryTokenValidator{role: tt.role}))
			request := httptest.NewRequest(http.MethodPut, "/api/v1/categories/1", strings.NewReader(`{"is_active":false}`))
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
			if catalogService.updateCalls != wantCalls {
				t.Fatalf("update calls = %d, want %d", catalogService.updateCalls, wantCalls)
			}
		})
	}
}

func TestCategoryHandlerListCategoriesPassesSupportedFilters(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantParent *int
		wantActive *bool
	}{
		{name: "parent", query: "parent_id=12", wantParent: intPointer(12)},
		{name: "active true", query: "is_active=true", wantActive: boolPointer(true)},
		{name: "active false", query: "is_active=false", wantActive: boolPointer(false)},
		{name: "combined", query: "parent_id=12&is_active=false", wantParent: intPointer(12), wantActive: boolPointer(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &categoryServiceStub{}
			responseRecorder := performCategoryListRequest(service, tt.query)

			if responseRecorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusOK)
			}
			if service.listCalls != 1 {
				t.Fatalf("service calls = %d, want 1", service.listCalls)
			}
			assertOptionalIntFilter(t, service.filters.ParentID, tt.wantParent)
			assertOptionalBoolFilter(t, service.filters.IsActive, tt.wantActive)
		})
	}
}

func TestCategoryHandlerListCategoriesPreservesUnfilteredPagination(t *testing.T) {
	service := &categoryServiceStub{}
	responseRecorder := performCategoryListRequest(service, "page=2&limit=25")

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusOK)
	}
	if service.filters.ParentID != nil || service.filters.IsActive != nil {
		t.Fatalf("filters = %+v, want empty filters", service.filters)
	}
	wantPagination := domain.Pagination{Page: 2, Limit: 25, Offset: 25}
	if service.pagination != wantPagination {
		t.Errorf("pagination = %+v, want %+v", service.pagination, wantPagination)
	}
}

func TestCategoryHandlerListCategoriesPreservesFirstRepeatedFilterValue(t *testing.T) {
	service := &categoryServiceStub{}
	responseRecorder := performCategoryListRequest(service, "parent_id=12&parent_id=24&is_active=false&is_active=true")

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusOK)
	}
	assertOptionalIntFilter(t, service.filters.ParentID, intPointer(12))
	assertOptionalBoolFilter(t, service.filters.IsActive, boolPointer(false))
}

func TestCategoryHandlerListCategoriesRejectsMalformedFilters(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "parent is not an integer", query: "parent_id=invalid"},
		{name: "parent is empty", query: "parent_id="},
		{name: "parent is zero", query: "parent_id=0"},
		{name: "parent is negative", query: "parent_id=-1"},
		{name: "parent overflows", query: "parent_id=999999999999999999999999999999999999"},
		{name: "active is not a boolean", query: "is_active=invalid"},
		{name: "active is empty", query: "is_active="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &categoryServiceStub{}
			responseRecorder := performCategoryListRequest(service, tt.query)

			if responseRecorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusBadRequest)
			}
			if service.listCalls != 0 {
				t.Fatalf("service calls = %d, want 0", service.listCalls)
			}

			var errorResponse response.ErrorResponse
			if err := json.Unmarshal(responseRecorder.Body.Bytes(), &errorResponse); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if errorResponse.Code != "invalid_request" {
				t.Errorf("error code = %q, want %q", errorResponse.Code, "invalid_request")
			}
		})
	}
}

func performCategoryListRequest(service domain.CategoryService, query string) *httptest.ResponseRecorder {
	handler := NewCategoryHandler(service)
	router := gin.New()
	router.GET("/api/v1/categories", handler.ListCategories)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/categories?"+query, nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	return responseRecorder
}

func assertOptionalIntFilter(t *testing.T, got, want *int) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("filter unexpectedly present with value %d", *got)
		}
		return
	}
	if got == nil {
		t.Fatal("filter is absent")
	}
	if *got != *want {
		t.Errorf("filter = %d, want %d", *got, *want)
	}
}

func assertOptionalBoolFilter(t *testing.T, got, want *bool) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("filter unexpectedly present with value %t", *got)
		}
		return
	}
	if got == nil {
		t.Fatal("filter is absent")
	}
	if *got != *want {
		t.Errorf("filter = %t, want %t", *got, *want)
	}
}

func intPointer(value int) *int    { return &value }
func boolPointer(value bool) *bool { return &value }

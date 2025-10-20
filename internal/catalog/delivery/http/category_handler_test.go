package http

import (
	"e-commerce-go/internal/catalog/domain"
	"e-commerce-go/internal/catalog/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestListCategories(t *testing.T) {
	mockService := &mocks.MockCategoryService{
		Categories: []domain.Category{{ID: 1, Name: "Tech"}},
		Total:      1,
	}

	handler := NewCategoryHandler(mockService)
	router := gin.Default()
	router.GET("/categories", handler.ListCategories)

	req, _ := http.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

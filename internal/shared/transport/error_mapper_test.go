package transport

import (
	"errors"
	"net/http"
	"testing"

	"e-commerce-go/internal/catalog/domain"
)

func TestHTTPErrorMapperClassifiesCatalogValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "product description", err: domain.ErrProductDescriptionRequired},
		{name: "SEO title", err: domain.ErrSeoTitle},
		{name: "SEO description", err: domain.ErrSeoDescription},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := HTTPErrorMapper(tt.err)
			if mapping.HTTPCode != http.StatusBadRequest || mapping.Code != "invalid_request" || mapping.LogLevel != LevelWarn {
				t.Fatalf("mapping = %+v, want 400 invalid_request WARN", mapping)
			}
		})
	}
}

func TestHTTPErrorMapperKeepsUnknownErrorsInternal(t *testing.T) {
	mapping := HTTPErrorMapper(errors.New("database unavailable"))
	if mapping.HTTPCode != http.StatusInternalServerError || mapping.Code != "internal_error" || mapping.LogLevel != LevelError {
		t.Fatalf("mapping = %+v, want 500 internal_error ERROR", mapping)
	}
}

package transport

import (
	"errors"
	"net/http"

	"e-commerce-go/internal/catalog/domain"
)

type HTTPErrorMapping struct {
	Code     string
	HTTPCode int
	LogLevel string
}

func HTTPErrorMapper(err error) HTTPErrorMapping {
	switch {
	case err == nil:
		return HTTPErrorMapping{"", http.StatusOK, "INFO"}

	// Category errors
	case errors.Is(err, domain.ErrCategoryNotFound):
		return HTTPErrorMapping{"not_found", http.StatusNotFound, "INFO"}
	case errors.Is(err, domain.ErrCategoryDescriptionRequired):
		return HTTPErrorMapping{"invalid_request", http.StatusBadRequest, "WARN"}

	// Product errors
	case errors.Is(err, domain.ErrInvalidProductID):
		return HTTPErrorMapping{"invalid_id", http.StatusBadRequest, "WARN"}
	case errors.Is(err, domain.ErrProductSlugExists):
		return HTTPErrorMapping{"conflict", http.StatusConflict, "WARN"}
	case errors.Is(err, domain.ErrProductNotFound):
		return HTTPErrorMapping{"not_found", http.StatusNotFound, "INFO"}

	default:
		return HTTPErrorMapping{"internal_error", http.StatusInternalServerError, "ERROR"}
	}
}

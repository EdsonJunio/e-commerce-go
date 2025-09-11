package http

import (
	"errors"
	"net/http"

	"e-commerce-go/internal/catalog/domain"
)

// CategoryHTTPCode maps domain errors to HTTP status codes and response codes.
// It ensures transport layer stays decoupled from domain logic.
func CategoryHTTPCode(err error) (string, int) {
	switch {
	case err == nil:
		return "", http.StatusOK

	case errors.Is(err, domain.ErrInvalidProductID):
		return "invalid_id", http.StatusBadRequest
	case errors.Is(err, domain.ErrProductNameRequired) || errors.Is(err, domain.ErrProductNameRequired) ||
		errors.Is(err, domain.ErrProductDescriptionRequired) || errors.Is(err, domain.ErrProductSlugRequired):
		return "invalid_request", http.StatusBadRequest
	case errors.Is(err, domain.ErrProductSlugExists):
		return "conflict", http.StatusConflict

	case errors.Is(err, domain.ErrProductNotFound):
		return "service_error", http.StatusNotFound

	default:
		return "internal_error", http.StatusInternalServerError
	}
}

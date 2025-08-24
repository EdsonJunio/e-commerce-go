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

	case errors.Is(err, domain.ErrInvalidID):
		return "invalid_id", http.StatusBadRequest
	case errors.Is(err, domain.ErrNameReq) || errors.Is(err, domain.ErrSlugReq) ||
		errors.Is(err, domain.ErrPriceInvalid) || errors.Is(err, domain.ErrSlugIsReq):
		return "invalid_request", http.StatusBadRequest
	case errors.Is(err, domain.ErrSlugExists):
		return "conflict", http.StatusConflict

	case errors.Is(err, domain.ErrNotFound):
		return "service_error", http.StatusNotFound

	default:
		return "internal_error", http.StatusInternalServerError
	}
}

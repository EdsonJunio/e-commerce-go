package http

import (
	"net/http"

	"e-commerce-go/internal/product/domain"
)

// ProductHTTPCode maps domain errors to HTTP status codes and response codes.
// It ensures transport layer stays decoupled from domain logic.
func ProductHTTPCode(err error) (string, int) {
	switch err {
	case nil:
		return "", http.StatusOK

	case domain.ErrInvalidID:
		return "invalid_id", http.StatusBadRequest
	case domain.ErrNameReq, domain.ErrSlugReq, domain.ErrPriceInvalid, domain.ErrSlugIsReq:
		return "invalid_request", http.StatusBadRequest
	case domain.ErrSlugExists:
		return "conflict", http.StatusConflict

	case domain.ErrNotFound:
		return "not_found", http.StatusNotFound

	default:
		return "internal_error", http.StatusInternalServerError
	}
}

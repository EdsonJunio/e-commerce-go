package transport

import (
	"errors"
	"net/http"

	"e-commerce-go/internal/catalog/domain"
)

type LogLevel string

const (
	LevelError LogLevel = "ERROR"
	LevelWarn  LogLevel = "WARN"
	LevelInfo  LogLevel = "INFO"
	LevelDebug LogLevel = "DEBUG"
)

type HTTPErrorMapping struct {
	Code     string
	HTTPCode int
	LogLevel LogLevel
}

var (
	mappingNotFound       = HTTPErrorMapping{"not_found", http.StatusNotFound, LevelInfo}
	mappingConflict       = HTTPErrorMapping{"conflict", http.StatusConflict, LevelWarn}
	mappingInvalidRequest = HTTPErrorMapping{"invalid_request", http.StatusBadRequest, LevelWarn}
	mappingInternal       = HTTPErrorMapping{"internal_error", http.StatusInternalServerError, LevelError}
)

func HTTPErrorMapper(err error) HTTPErrorMapping {
	if err == nil {
		return HTTPErrorMapping{"", http.StatusOK, LevelInfo}
	}

	switch {
	// 404 Not Found Group
	case errors.Is(err, domain.ErrCategoryNotFound),
		errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, domain.ErrParentCategoryNotFound):
		return mappingNotFound

	// 409 Conflict Group
	case errors.Is(err, domain.ErrCategorySlugExists),
		errors.Is(err, domain.ErrProductSlugExists):
		return mappingConflict

	// 400 Bad Request Group (Validation)
	case errors.Is(err, domain.ErrCategoryDescriptionRequired),
		errors.Is(err, domain.ErrInvalidProductID),
		errors.Is(err, domain.ErrInvalidCategoryReference),
		errors.Is(err, domain.ErrProductSlugRequired),
		errors.Is(err, domain.ErrProductNameRequired),
		errors.Is(err, domain.ErrInvalidCategoryID),
		errors.Is(err, domain.ErrCategoryNameRequired),
		errors.Is(err, domain.ErrCategorySlugRequired):
		return mappingInvalidRequest

	// Default 500
	default:
		return mappingInternal
	}
}

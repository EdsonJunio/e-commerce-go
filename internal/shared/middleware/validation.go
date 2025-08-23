package middleware

import (
	"net/http"
	"regexp"

	"e-commerce-go/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidations registers custom validation rules.
func RegisterCustomValidations() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
			matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, fl.Field().String())
			return matched
		})
	}
}

// ErrorHandler is a middleware for handling validation and internal errors.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			switch e := err.Err.(type) {
			case validator.ValidationErrors:
				var errors []string
				for _, fieldErr := range e {
					errors = append(errors, fieldErrorToMessage(fieldErr))
				}
				c.JSON(http.StatusBadRequest, gin.H{
					"error":  "validation_error",
					"fields": errors,
				})
				return
			default:
				response.AbortWithError(c, http.StatusInternalServerError, "internal_error", e.Error())
			}
		}
	}
}

// fieldErrorToMessage converts a validation error into a human-readable message.
func fieldErrorToMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fieldErr.Field() + " is required"
	case "min":
		return fieldErr.Field() + " must have at least " + fieldErr.Param() + " characters"
	case "max":
		return fieldErr.Field() + " must have at most " + fieldErr.Param() + " characters"
	case "email":
		return "invalid email format"
	case "slug":
		return fieldErr.Field() + " must contain only lowercase letters, numbers, and hyphens"
	default:
		return fieldErr.Field() + " is invalid"
	}
}

package middleware

import (
	"net/http"
	"regexp"

	"e-commerce-go/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidations registra validações personalizadas
func RegisterCustomValidations() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Validador de slug (letras minúsculas, números e hífens)
		_ = v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
			matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, fl.Field().String())
			return matched
		})
	}
}

// ErrorHandler é um middleware para tratamento de erros de validação
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

// fieldErrorToMessage converte um erro de validação em uma mensagem legível
func fieldErrorToMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fieldErr.Field() + " é obrigatório"
	case "min":
		return fieldErr.Field() + " deve ter pelo menos " + fieldErr.Param() + " caracteres"
	case "max":
		return fieldErr.Field() + " deve ter no máximo " + fieldErr.Param() + " caracteres"
	case "email":
		return "E-mail inválido"
	case "slug":
		return fieldErr.Field() + " deve conter apenas letras minúsculas, números e hífens"
	default:
		return fieldErr.Field() + " é inválido"
	}
}

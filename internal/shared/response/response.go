package response

import (
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response.
// @Description Estrutura padrão para retorno de erros da API
type ErrorResponse struct {
	// Código interno do erro (ex: "INVALID_INPUT", "NOT_FOUND")
	Code string `json:"code" example:"INVALID_INPUT"`
	// Mensagem descritiva para o usuário
	Message string `json:"message" example:"The field 'email' is required"`
}

// PaginatedResponse represents a paginated response.
// @Description Estrutura padrão para listas paginadas
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contains pagination details.
type Pagination struct {
	Total   int64  `json:"total" example:"100"`
	Page    int    `json:"page" example:"1"`
	Limit   int    `json:"limit" example:"10"`
	Pages   int    `json:"pages" example:"10"`
	HasMore bool   `json:"has_more" example:"true"`
	NextURL string `json:"next_url,omitempty" example:"/api/v1/resource?page=2&limit=10"`
	PrevURL string `json:"prev_url,omitempty" example:""`
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// NewSuccessResponse creates a new success response.
func NewSuccessResponse(data interface{}) gin.H {
	return gin.H{
		"data": data,
	}
}

// NewPaginatedResponse creates a new paginated response.
func NewPaginatedResponse(data interface{}, total int64, page, limit int, basePath string) PaginatedResponse {
	pages := int((total + int64(limit) - 1) / int64(limit))
	hasMore := page*limit < int(total)

	var nextURL, prevURL string

	if hasMore {
		nextURL = buildPaginationURL(basePath, page+1, limit)
	}

	if page > 1 {
		prevURL = buildPaginationURL(basePath, page-1, limit)
	}

	return PaginatedResponse{
		Data: data,
		Pagination: Pagination{
			Total:   total,
			Page:    page,
			Limit:   limit,
			Pages:   pages,
			HasMore: hasMore,
			NextURL: nextURL,
			PrevURL: prevURL,
		},
	}
}

// buildPaginationURL builds the pagination URL.
func buildPaginationURL(basePath string, page, limit int) string {
	return path.Join("/", basePath) + "?page=" + strconv.Itoa(page) + "&limit=" + strconv.Itoa(limit)
}

// AbortWithError aborts the request with an error response.
func AbortWithError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, NewErrorResponse(code, message))
}

// Success sends a success response.
func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, NewSuccessResponse(data))
}

// SuccessPaginated sends a paginated success response.
func SuccessPaginated(c *gin.Context, status int, data interface{}, total int64, page, limit int) {
	c.JSON(status, NewPaginatedResponse(data, total, page, limit, c.Request.URL.Path))
}

// Error sends an error response.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, NewErrorResponse(code, message))
}

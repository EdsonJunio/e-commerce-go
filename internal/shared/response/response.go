package response

import (
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ErrorResponse representa uma resposta de erro padronizada
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PaginatedResponse representa uma resposta paginada
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contém informações de paginação
type Pagination struct {
	Total   int64  `json:"total"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Pages   int    `json:"pages"`
	HasMore bool   `json:"has_more"`
	NextURL string `json:"next_url,omitempty"`
	PrevURL string `json:"prev_url,omitempty"`
}

// NewErrorResponse cria uma nova resposta de erro
func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// NewSuccessResponse cria uma nova resposta de sucesso
func NewSuccessResponse(data interface{}) gin.H {
	return gin.H{
		"data": data,
	}
}

// NewPaginatedResponse cria uma nova resposta paginada
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

// buildPaginationURL constrói a URL para paginação
func buildPaginationURL(basePath string, page, limit int) string {
	return path.Join("/", basePath) + "?page=" + strconv.Itoa(page) + "&limit=" + strconv.Itoa(limit)
}

// AbortWithError aborta a requisição com uma mensagem de erro
func AbortWithError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, NewErrorResponse(code, message))
}

// Success envia uma resposta de sucesso
func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, NewSuccessResponse(data))
}

// SuccessPaginated envia uma resposta de sucesso paginada
func SuccessPaginated(c *gin.Context, status int, data interface{}, total int64, page, limit int) {
	c.JSON(status, NewPaginatedResponse(data, total, page, limit, c.Request.URL.Path))
}

// Error envia uma resposta de erro
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, NewErrorResponse(code, message))
}

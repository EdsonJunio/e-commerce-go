package http

import (
	"e-commerce-go/internal/identity/domain"
	"e-commerce-go/internal/shared/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@gmail.com"`
	Password string `json:"password" binding:"required" example:"hash123"`
}

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1Ni..."`
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", h.Login)
	}
}

// Login godoc
// @Summary      User Login
// @Description  Authenticates a user and returns a JWT Token
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login Credentials"
// @Success      200  {object}  LoginResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid_request",
			err.Error(),
		)

		return
	}

	token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		response.Error(
			c,
			http.StatusUnauthorized,
			"unauthorized",
			"Email or password incorrect",
		)

		return
	}

	response.Success(c, http.StatusOK, LoginResponse{Token: token})
}

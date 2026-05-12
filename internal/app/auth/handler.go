package auth

import (
	"errors"
	"net/http"

	"backend-go/internal/core/domainerr"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Login realiza a autenticação do usuário
// @Summary Realizar Login
// @Description Autentica o usuário com email e senha, retornando tokens JWT e dados do usuário.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credenciais de acesso"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string "Dados inválidos"
// @Failure 401 {object} map[string]string "Credenciais inválidas"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	res, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Me retorna os dados do usuário autenticado
// @Summary Obter dados do usuário atual
// @Description Retorna informações do usuário logado baseado no token JWT.
// @Tags Auth
// @Produce json
// @Security Bearer
// @Success 200 {object} LoginResponse
// @Failure 401 {object} map[string]string "Não autorizado"
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	email, exists := c.Get("userEmail")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não identificado"})
		return
	}

	res, err := h.service.GetMe(c.Request.Context(), email.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Refresh renova o token de acesso
// @Summary Renovar token
// @Description Gera um novo access token a partir de um refresh token válido.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string "Dados inválidos"
// @Failure 401 {object} map[string]string "Refresh token inválido ou expirado"
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	res, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "erro interno do servidor"

	switch {
	case errors.Is(err, domainerr.ErrUserNotFound), errors.Is(err, domainerr.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		message = err.Error()
	case errors.Is(err, domainerr.ErrAccountDisabled):
		status = http.StatusForbidden
		message = err.Error()
	case errors.Is(err, domainerr.ErrAuthNotConfigured):
		status = http.StatusUnprocessableEntity
		message = err.Error()
	}

	c.JSON(status, gin.H{"error": message})
}

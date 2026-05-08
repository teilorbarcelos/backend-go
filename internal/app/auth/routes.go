package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registra as rotas que não exigem autenticação
func RegisterPublicRoutes(rg *gin.RouterGroup, h *AuthHandler) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
	}
}

// RegisterProtectedRoutes registra as rotas que exigem autenticação
func RegisterProtectedRoutes(rg *gin.RouterGroup, h *AuthHandler) {
	rg.GET("/auth/me", h.Me)
}

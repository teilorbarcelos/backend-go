package auth

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"backend-go/internal/core/repository"
)

// RegisterRoutes inicializa o módulo e registra as rotas públicas e protegidas
func RegisterRoutes(publicRG *gin.RouterGroup, protectedRG *gin.RouterGroup, db *gorm.DB) {
	repo := repository.NewAuthRepository(db)
	svc := NewAuthService(repo)
	h := NewAuthHandler(svc)

	// Rotas Públicas
	authGroup := publicRG.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
	}

	// Rotas Protegidas
	protectedRG.GET("/auth/me", h.Me)
}

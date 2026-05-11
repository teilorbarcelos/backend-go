package auth

import (
	"backend-go/internal/core/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(publicRG *gin.RouterGroup, protectedRG *gin.RouterGroup, db *gorm.DB) {
	repo := repository.NewAuthRepository(db)
	svc := NewAuthService(repo)
	h := NewAuthHandler(svc)

	authGroup := publicRG.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
	}
	protectedRG.GET("/auth/me", h.Me)
}

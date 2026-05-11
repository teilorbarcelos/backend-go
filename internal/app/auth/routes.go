package auth

import (
	"backend-go/internal/infra/session"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(publicRG *gin.RouterGroup, protectedRG *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	sm := session.NewSessionManager()
	svc := NewService(repo, sm)
	h := NewHandler(svc)

	authGroup := publicRG.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
	}
	protectedRG.GET("/auth/me", h.Me)
}

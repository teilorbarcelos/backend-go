package user

import (
	"github.com/gin-gonic/gin"
	"github.com/teilorbarcelos/backend-go/internal/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *UserHandler) {
	userRoutes := rg.Group("/user")
	{
		userRoutes.GET("/:id", middleware.CheckPermission("user", "view"), h.GetByID)
		userRoutes.GET("", middleware.CheckPermission("user", "view"), h.List)
		userRoutes.GET("/all", middleware.CheckPermission("user", "view"), h.ListAll)
		userRoutes.POST("", middleware.CheckPermission("user", "create"), h.Create)
		userRoutes.PUT("/:id", middleware.CheckPermission("user", "create"), h.Update)
		userRoutes.DELETE("/:id", middleware.CheckPermission("user", "delete"), h.Delete)
		userRoutes.PATCH("/:id/status", middleware.CheckPermission("user", "activate"), h.SetStatus)
	}
}

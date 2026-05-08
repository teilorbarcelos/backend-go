package role

import (
	"github.com/gin-gonic/gin"
	"github.com/teilorbarcelos/backend-go/internal/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *RoleHandler) {
	roleRoutes := rg.Group("/role")
	{
		roleRoutes.GET("/features", middleware.CheckPermission("role", "view"), h.ListFeatures)
		roleRoutes.GET("/:id", middleware.CheckPermission("role", "view"), h.GetByID)
		roleRoutes.GET("", middleware.CheckPermission("role", "view"), h.List)
		roleRoutes.GET("/all", middleware.CheckPermission("role", "view"), h.ListAll)
		roleRoutes.POST("", middleware.CheckPermission("role", "create"), h.Create)
		roleRoutes.PUT("/:id", middleware.CheckPermission("role", "create"), h.Update)
		roleRoutes.DELETE("/:id", middleware.CheckPermission("role", "delete"), h.Delete)
		roleRoutes.PATCH("/:id/status", middleware.CheckPermission("role", "activate"), h.SetStatus)
	}
}

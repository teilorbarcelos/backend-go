package middleware

import (
	"net/http"

	"backend-go/pkg/security"

	"github.com/gin-gonic/gin"
)

func hasPermission(userPerms []security.Permission, feature string, action string) bool {
	for _, p := range userPerms {
		if p.Feature == feature {
			switch action {
			case "view":
				return p.View
			case "create":
				return p.Create
			case "delete":
				return p.Delete
			case "activate":
				return p.Activate
			}
		}
	}
	return false
}

func CheckPermission(feature string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, roleExists := c.Get("userRoleID")
		if roleExists && roleID.(string) == "administrator" {
			c.Next()
			return
		}

		permissions, exists := c.Get("userPermissions")
		if !exists || permissions == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permissões não encontradas no contexto. Faça login novamente."})
			c.Abort()
			return
		}

		userPerms := permissions.([]security.Permission)

		if !hasPermission(userPerms, feature, action) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Você não tem permissão para realizar esta ação",
				"details": gin.H{
					"feature": feature,
					"action":  action,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

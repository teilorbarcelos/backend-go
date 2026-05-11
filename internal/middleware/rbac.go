package middleware

import (
	"net/http"

	"backend-go/pkg/security"

	"github.com/gin-gonic/gin"
)

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

		var hasPerm bool
		for _, p := range userPerms {
			if p.Feature == feature {
				switch action {
				case "view":
					hasPerm = p.View
				case "create":
					hasPerm = p.Create
				case "delete":
					hasPerm = p.Delete
				case "activate":
					hasPerm = p.Activate
				}
				break
			}
		}

		if !hasPerm {
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

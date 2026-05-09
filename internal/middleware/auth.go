package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"backend-go/pkg/cache"
	"backend-go/pkg/security"
)

// Authenticate é um middleware que valida o token JWT no header Authorization.
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header é obrigatório"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de autorização inválido. Use 'Bearer <token>'"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := security.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido ou expirado"})
			c.Abort()
			return
		}

		// Validar sessão no Redis (conforme backend-node)
		tokenHash := security.SHA256(tokenString)
		sessionKey := fmt.Sprintf("session:role:%s:user:%s:access:%s", claims.RoleID, claims.UserID, tokenHash)
		
		exists, err := cache.RedisClient.Exists(c.Request.Context(), sessionKey).Result()
		if err != nil || exists == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Sessão inválida ou expirada"})
			c.Abort()
			return
		}

		// Injeta os dados do usuário no contexto para uso posterior nos handlers e audit hooks
		ctx := context.WithValue(c.Request.Context(), "userID", claims.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRoleID", claims.RoleID)
		c.Set("userPermissions", claims.Permissions)

		c.Next()
	}
}

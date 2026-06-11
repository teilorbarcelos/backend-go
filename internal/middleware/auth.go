package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"backend-go/pkg/cache"
	"backend-go/pkg/security"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const middlewareSessionVerKey = "session:ver:%s"

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
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Token inválido: %v", err)})
			c.Abort()
			return
		}

		fmt.Printf("[AUTH DEBUG] UserID=%s Email=%s RoleID=%s SessionVersion=%d Permissions=%d\n",
			claims.UserID, claims.Email, claims.RoleID, claims.SessionVersion, len(claims.Permissions))

		storedVersion, err := cache.RedisClient.Get(c.Request.Context(), fmt.Sprintf(middlewareSessionVerKey, claims.UserID)).Int()
		if errors.Is(err, redis.Nil) {
			storedVersion = 0
		} else if err != nil {
			errMsg := fmt.Sprintf("Redis error: %v", err)
			fmt.Printf("[AUTH DEBUG] %s\n", errMsg)
			c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
			c.Abort()
			return
		}
		if storedVersion != claims.SessionVersion {
			fmt.Printf("[AUTH DEBUG] Version mismatch: stored=%d claims=%d\n", storedVersion, claims.SessionVersion)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "UnauthorizedError"})
			c.Abort()
			return
		}

		ctx := context.WithValue(c.Request.Context(), "userID", claims.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRoleID", claims.RoleID)
		c.Set("userPermissions", claims.Permissions)
		if len(claims.Permissions) > 0 {
			c.Set("userPermissionsBitset", security.CompilePermissions(claims.Permissions))
		}

		c.Next()
	}
}

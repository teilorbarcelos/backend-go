package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/teilorbarcelos/backend-go/pkg/cache"
	"github.com/teilorbarcelos/backend-go/pkg/config"
	"github.com/teilorbarcelos/backend-go/pkg/security"
)

func TestAuthenticate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config.LoadConfig()
	if cache.RedisClient == nil {
		cache.ConnectRedis()
	}

	r := gin.New()
	r.Use(Authenticate())
	r.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	// 1. Sem header
	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 2. Token Inválido (Formato)
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 2.1 Token Inválido (JWT)
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 3. Token Válido mas sem sessão no Redis
	token, _ := security.GenerateToken("user-123", "user@test.com", "role-admin", []security.Permission{{Feature: "user", View: true}})
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 4. Token Válido com sessão no Redis
	tokenHash := security.SHA256(token)
	sessionKey := fmt.Sprintf("session:role:role-admin:user:user-123:access:%s", tokenHash)
	cache.RedisClient.Set(context.Background(), sessionKey, "active", time.Minute)
	defer cache.RedisClient.Del(context.Background(), sessionKey)

	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user-123")
}

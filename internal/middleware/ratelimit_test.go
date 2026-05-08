package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/teilorbarcelos/backend-go/pkg/cache"
	"github.com/teilorbarcelos/backend-go/pkg/config"
)

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config.LoadConfig()
	// Mock Redis se necessário, mas como já temos o cache.ConnectRedis nos outros testes, 
	// assumimos que o RedisClient está disponível (ou usamos um dummy se falhar)
	if cache.RedisClient == nil {
		cache.ConnectRedis()
	}

	r := gin.New()
	
	// Salva ambiente original e restaura depois
	origEnv := config.AppConfig.Environment
	origMax := config.AppConfig.RateLimitMax
	origWindow := config.AppConfig.RateLimitWindow
	
	config.AppConfig.Environment = "development" // Força execução da lógica
	config.AppConfig.RateLimitMax = 3
	config.AppConfig.RateLimitWindow = "10s"
	
	defer func() {
		config.AppConfig.Environment = origEnv
		config.AppConfig.RateLimitMax = origMax
		config.AppConfig.RateLimitWindow = origWindow
	}()

	r.Use(RateLimitMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Teste de Bypass no ambiente 'test'
	config.AppConfig.Environment = "test"
	reqTest, _ := http.NewRequest("GET", "/test", nil)
	wTest := httptest.NewRecorder()
	r.ServeHTTP(wTest, reqTest)
	assert.Equal(t, http.StatusOK, wTest.Code)
	config.AppConfig.Environment = "development"

	ctx := context.Background()
	key := "ratelimit:ip:127.0.0.1"
	cache.RedisClient.Del(ctx, key)

	// Requisição 1: OK (Cobre o ramo current == 0 e pipe.Expire)
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Teste de Erro de Parse na Janela (Cobre o branch de erro de ParseDuration)
	config.AppConfig.RateLimitWindow = "invalid-duration"
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	config.AppConfig.RateLimitWindow = "10s"

	// Requisição 2: OK
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Requisição 3: Limite Excedido
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Teste de Rate Limit usando UserID (Cobre o branch 'if exists')
	rUser := gin.New()
	rUser.Use(func(c *gin.Context) {
		c.Set("userID", "user-456")
		c.Next()
	})
	rUser.Use(RateLimitMiddleware())
	rUser.GET("/test-user", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	
	cache.RedisClient.Del(ctx, "ratelimit:user:user-456")
	reqUser, _ := http.NewRequest("GET", "/test-user", nil)
	wUser := httptest.NewRecorder()
	rUser.ServeHTTP(wUser, reqUser)
	assert.Equal(t, http.StatusOK, wUser.Code)
}

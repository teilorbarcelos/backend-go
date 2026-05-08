package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/teilorbarcelos/backend-go/pkg/cache"
	"github.com/teilorbarcelos/backend-go/pkg/config"
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.AppConfig.Environment == "test" {
			c.Next()
			return
		}

		// O rate limit no Node usava o ID do usuário (AuthPayload) ou IP como fallback
		userID, exists := c.Get("userID")
		var key string
		if exists {
			key = "ratelimit:user:" + userID.(string)
		} else {
			key = "ratelimit:ip:" + c.ClientIP()
		}

		windowStr := config.AppConfig.RateLimitWindow // ex: "1m"
		windowDuration, err := time.ParseDuration(windowStr)
		if err != nil {
			windowDuration = time.Minute
		}

		maxRequests := int64(config.AppConfig.RateLimitMax)

		ctx := context.Background()
		
		// Usar um INCR e EXPIRE básicos
		current, err := cache.RedisClient.Get(ctx, key).Int64()
		if err == nil && current >= maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "Você excedeu o limite de requisições. Tente novamente em breve.",
			})
			return
		}

		pipe := cache.RedisClient.Pipeline()
		pipe.Incr(ctx, key)
		if current == 0 { // Set expire only on first request in the window
			pipe.Expire(ctx, key, windowDuration)
		}
		_, _ = pipe.Exec(ctx)

		c.Header("X-RateLimit-Limit", strconv.FormatInt(maxRequests, 10))
		// Observação: para retornar o "Remaining" de forma mais exata seria necessário LUA scripts ou bibliotecas específicas. 
		// Manteremos simples como no Node para MVP.

		c.Next()
	}
}

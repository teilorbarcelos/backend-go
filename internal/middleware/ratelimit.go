package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"backend-go/pkg/cache"
	"backend-go/pkg/config"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.AppConfig.Environment == "test" {
			c.Next()
			return
		}

		userID, exists := c.Get("userID")
		var key string
		if exists {
			key = "ratelimit:user:" + userID.(string)
		} else {
			key = "ratelimit:ip:" + c.ClientIP()
		}

		windowStr := config.AppConfig.RateLimitWindow
		windowDuration, err := time.ParseDuration(windowStr)
		if err != nil {
			windowDuration = time.Minute
		}

		maxRequests := int64(config.AppConfig.RateLimitMax)

		ctx := context.Background()

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
		if current == 0 {
			pipe.Expire(ctx, key, windowDuration)
		}
		_, _ = pipe.Exec(ctx)

		c.Header("X-RateLimit-Limit", strconv.FormatInt(maxRequests, 10))

		c.Next()
	}
}

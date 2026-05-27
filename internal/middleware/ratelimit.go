package middleware

import (
	"net/http"
	"strconv"
	"time"

	"backend-go/pkg/cache"
	"backend-go/pkg/config"

	"github.com/gin-gonic/gin"
)

func getRateLimitKey(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if exists {
		return "ratelimit:user:" + userID.(string)
	}
	return "ratelimit:ip:" + c.ClientIP()
}

func getRateLimitConfig() (time.Duration, int64) {
	windowStr := config.AppConfig.RateLimitWindow
	windowDuration, err := time.ParseDuration(windowStr)
	if err != nil {
		windowDuration = time.Minute
	}
	maxRequests := int64(config.AppConfig.RateLimitMax)
	return windowDuration, maxRequests
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.AppConfig.Environment == "test" {
			c.Next()
			return
		}

		key := getRateLimitKey(c)
		windowDuration, maxRequests := getRateLimitConfig()

		ctx := c.Request.Context()

		pipe := cache.RedisClient.Pipeline()
		incr := pipe.Incr(ctx, key)
		ttl := pipe.TTL(ctx, key)
		_, err := pipe.Exec(ctx)

		if err != nil && err != ctx.Err() {
			c.Next()
			return
		}

		currentCount := incr.Val()

		if currentCount == 1 || ttl.Val() < 0 {
			cache.RedisClient.Expire(ctx, key, windowDuration)
			ttl = cache.RedisClient.TTL(ctx, key)
		}

		remaining := maxRequests - currentCount
		if remaining < 0 {
			remaining = 0
		}

		resetInSeconds := int64(ttl.Val().Seconds())

		c.Header("X-RateLimit-Limit", strconv.FormatInt(maxRequests, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetInSeconds, 10))

		if currentCount > maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "Você excedeu o limite de requisições. Tente novamente em breve.",
				"details": gin.H{
					"limit":     maxRequests,
					"remaining": 0,
					"reset_in":  resetInSeconds,
				},
			})
			return
		}

		c.Next()
	}
}

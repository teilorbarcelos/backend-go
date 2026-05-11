package main

import (
	"log"

	"backend-go/internal/app/auth"
	"backend-go/internal/app/product"
	"backend-go/internal/app/role"
	"backend-go/internal/app/user"
	"backend-go/internal/infra/session"
	"backend-go/internal/middleware"
	"backend-go/pkg/cache"
	"backend-go/pkg/config"
	"backend-go/pkg/database"
	"backend-go/pkg/messaging"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()

	if config.AppConfig.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	database.ConnectDB()
	cache.ConnectRedis()
	messaging.ConnectRabbitMQ()
	defer messaging.RabbitConn.Close()
	defer messaging.RabbitChannel.Close()
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.RateLimitMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "ok",
			"environment": config.AppConfig.Environment,
		})
	})

	sessionMgr := session.NewSessionManager()

	v1 := r.Group("/v1")
	{
		protected := v1.Group("/")
		protected.Use(middleware.Authenticate())
		auth.RegisterRoutes(v1, protected, database.DB)
		user.RegisterRoutes(protected, database.DB, sessionMgr)
		role.RegisterRoutes(protected, database.DB, sessionMgr)
		product.RegisterRoutes(protected, database.DB)
	}

	addr := config.AppConfig.Host + ":" + config.AppConfig.Port
	log.Printf("Iniciando servidor em http://%s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

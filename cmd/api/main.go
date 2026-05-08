package main

import (
	"log"

	"github.com/gin-gonic/gin"
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
)

func main() {
	// 1. Carrega configurações (.env)
	config.LoadConfig()

	// 2. Ajusta modo do Gin baseado no Environment
	if config.AppConfig.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3. Inicializa conexões com Infraestrutura
	database.ConnectDB()
	cache.ConnectRedis()
	messaging.ConnectRabbitMQ()
	defer messaging.RabbitConn.Close()
	defer messaging.RabbitChannel.Close()

	// 4. Configura o router do Gin
	r := gin.Default()

	// Middlewares Globais
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimitMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "ok",
			"environment": config.AppConfig.Environment,
		})
	})

	// Repositories, Services e Handlers (Modularizado)
	sessionMgr := session.NewSessionManager()

	// Grupo V1
	v1 := r.Group("/v1")
	{
		// Grupo Protegido (Autenticação)
		protected := v1.Group("/")
		protected.Use(middleware.Authenticate())

		// Registro dos Módulos
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

package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/teilorbarcelos/backend-go/internal/app/auth"
	"github.com/teilorbarcelos/backend-go/internal/app/product"
	"github.com/teilorbarcelos/backend-go/internal/app/role"
	"github.com/teilorbarcelos/backend-go/internal/app/user"
	"github.com/teilorbarcelos/backend-go/internal/core/repository"
	"github.com/teilorbarcelos/backend-go/internal/infra/session"
	"github.com/teilorbarcelos/backend-go/internal/middleware"
	"github.com/teilorbarcelos/backend-go/pkg/cache"
	"github.com/teilorbarcelos/backend-go/pkg/config"
	"github.com/teilorbarcelos/backend-go/pkg/database"
	"github.com/teilorbarcelos/backend-go/pkg/messaging"
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

	// Repositories, Services e Handlers
	sessionMgr := session.NewSessionManager()

	authRepo := repository.NewAuthRepository(database.DB)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	roleRepo := role.NewRoleRepository(database.DB)
	roleService := role.NewRoleService(roleRepo, sessionMgr)
	roleHandler := role.NewRoleHandler(roleService)

	userRepo := user.NewUserRepository(database.DB)
	userService := user.NewUserService(userRepo, sessionMgr)
	userHandler := user.NewUserHandler(userService)

	productRepo := product.NewProductRepository(database.DB)
	productService := product.NewProductService(productRepo)
	productHandler := product.NewProductHandler(productService)

	// Grupo V1
	v1 := r.Group("/v1")
	{
		// Rotas Públicas
		auth.RegisterPublicRoutes(v1, authHandler)

		// Rotas Privadas (Protegidas por Autenticação)
		protected := v1.Group("/")
		protected.Use(middleware.Authenticate())
		{
			auth.RegisterProtectedRoutes(protected, authHandler)
			user.RegisterRoutes(protected, userHandler)
			role.RegisterRoutes(protected, roleHandler)
			product.RegisterRoutes(protected, productHandler)
		}
	}

	addr := config.AppConfig.Host + ":" + config.AppConfig.Port
	log.Printf("Iniciando servidor em http://%s", addr)
	
	if err := r.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

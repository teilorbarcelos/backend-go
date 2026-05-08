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
		// Públicos
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
		}

		// Privados (Protegidos por Autenticação)
		protected := v1.Group("/")
		protected.Use(middleware.Authenticate())
		{
			// Auth (Me)
			protected.GET("/auth/me", authHandler.Me)

			// User
			userRoutes := protected.Group("/user")
			{
				userRoutes.GET("/:id", middleware.CheckPermission("user", "view"), userHandler.GetByID)
				userRoutes.GET("", middleware.CheckPermission("user", "view"), userHandler.List)
				userRoutes.GET("/all", middleware.CheckPermission("user", "view"), userHandler.ListAll)
				userRoutes.POST("", middleware.CheckPermission("user", "create"), userHandler.Create)
				userRoutes.PUT("/:id", middleware.CheckPermission("user", "create"), userHandler.Update)
				userRoutes.DELETE("/:id", middleware.CheckPermission("user", "delete"), userHandler.Delete)
				userRoutes.PATCH("/:id/status", middleware.CheckPermission("user", "activate"), userHandler.SetStatus)
			}

			// Role
			roleRoutes := protected.Group("/role")
			{
				roleRoutes.GET("/features", middleware.CheckPermission("role", "view"), roleHandler.ListFeatures)
				roleRoutes.GET("/:id", middleware.CheckPermission("role", "view"), roleHandler.GetByID)
				roleRoutes.GET("", middleware.CheckPermission("role", "view"), roleHandler.List)
				roleRoutes.GET("/all", middleware.CheckPermission("role", "view"), roleHandler.ListAll)
				roleRoutes.POST("", middleware.CheckPermission("role", "create"), roleHandler.Create)
				roleRoutes.PUT("/:id", middleware.CheckPermission("role", "create"), roleHandler.Update)
				roleRoutes.DELETE("/:id", middleware.CheckPermission("role", "delete"), roleHandler.Delete)
				roleRoutes.PATCH("/:id/status", middleware.CheckPermission("role", "activate"), roleHandler.SetStatus)
			}

			// Product
			productRoutes := protected.Group("/product")
			{
				productRoutes.GET("/:id", middleware.CheckPermission("product", "view"), productHandler.GetByID)
				productRoutes.GET("", middleware.CheckPermission("product", "view"), productHandler.List)
				productRoutes.GET("/all", middleware.CheckPermission("product", "view"), productHandler.ListAll)
				productRoutes.POST("", middleware.CheckPermission("product", "create"), productHandler.Create)
				productRoutes.PUT("/:id", middleware.CheckPermission("product", "create"), productHandler.Update)
				productRoutes.DELETE("/:id", middleware.CheckPermission("product", "delete"), productHandler.Delete)
				productRoutes.PATCH("/:id/status", middleware.CheckPermission("product", "activate"), productHandler.SetStatus)
			}
		}
	}

	addr := config.AppConfig.Host + ":" + config.AppConfig.Port
	log.Printf("Iniciando servidor em http://%s", addr)
	
	if err := r.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

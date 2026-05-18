package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-go/internal/app/auth"
	"backend-go/internal/app/debug"
	"backend-go/internal/app/product"
	"backend-go/internal/app/role"
	"backend-go/internal/app/user"
	"backend-go/internal/core/audit"
	"backend-go/internal/infra/session"
	"backend-go/internal/middleware"
	_ "backend-go/docs"
	"backend-go/pkg/cache"
	"backend-go/pkg/config"
	"backend-go/pkg/database"
	"backend-go/pkg/logger"
	"backend-go/pkg/messaging"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Backend Go API
// @version 1.0
// @description API modular em Go com Gin e Swagger.
// @termsOfService http://swagger.io/terms/

// @contact.name Suporte API
// @contact.url http://www.swagger.io/support
// @contact.email suporte@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8888
// @BasePath /v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

func main() {
	config.LoadConfig()
	logger.InitLogger(config.AppConfig.Environment)

	if config.AppConfig.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	database.ConnectDB()
	audit.RegisterAuditHooks(database.DB)
	cache.ConnectRedis()
	messaging.ConnectRabbitMQ()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorLogger())
	r.Use(middleware.Metrics())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimitMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "ok",
			"environment": config.AppConfig.Environment,
		})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	sessionMgr := session.NewSessionManager()

	v1 := r.Group("/v1")
	{
		v1.GET("/docs", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/v1/docs/index.html")
		})
		v1.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		protected := v1.Group("/")
		protected.Use(middleware.Authenticate())
		auth.RegisterRoutes(v1, protected, database.DB)
		user.RegisterRoutes(protected, database.DB, sessionMgr)
		role.RegisterRoutes(protected, database.DB, sessionMgr)
		product.RegisterRoutes(protected, database.DB)
		debug.RegisterRoutes(v1)
	}

	addr := config.AppConfig.Host + ":" + config.AppConfig.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logger.Log.Sugar().Infof("Iniciando servidor em http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Sugar().Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Encerrando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Sugar().Fatalf("Forçar encerramento do servidor: %v", err)
	}

	logger.Info("Limpando recursos...")
	if messaging.RabbitConn != nil {
		messaging.RabbitConn.Close()
	}

	logger.Info("Servidor finalizado com sucesso.")
}

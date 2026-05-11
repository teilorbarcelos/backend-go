package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("Iniciando servidor em http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Encerrando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forçar encerramento do servidor: %v", err)
	}

	log.Println("Limpando recursos...")
	if messaging.RabbitConn != nil {
		messaging.RabbitConn.Close()
	}

	log.Println("Servidor finalizado com sucesso.")
}

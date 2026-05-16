package config

import (
	"backend-go/pkg/logger"

	"github.com/spf13/viper"
)

type Config struct {
	Environment       string `mapstructure:"ENVIRONMENT"`
	Port              string `mapstructure:"PORT"`
	Host              string `mapstructure:"HOST"`
	DBUrl             string `mapstructure:"DATABASE_URL"`
	RedisUrl          string `mapstructure:"REDIS_URL"`
	RabbitMQUrl       string `mapstructure:"RABBITMQ_URL"`
	JWTSecret         string `mapstructure:"JWT_SECRET"`
	RateLimitMax      int    `mapstructure:"RATE_LIMIT_MAX"`
	RateLimitWindow   string `mapstructure:"RATE_LIMIT_WINDOW"`
	FirstUserEmail    string `mapstructure:"FIRST_USER"`
	FirstUserPassword string `mapstructure:"FIRST_PASSWORD"`
	LogLevel          string `mapstructure:"LOG_LEVEL"`
	PdfServiceUrl     string `mapstructure:"PDF_SERVICE_URL"`
}

var AppConfig Config

var (
	logFatalf      = logger.Fatalf
	viperUnmarshal = viper.Unmarshal
)

func LoadConfig() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("PORT", "3000")
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("RATE_LIMIT_MAX", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW", "1m")
	viper.SetDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	viper.SetDefault("FIRST_USER", "admin@email.com")
	viper.SetDefault("FIRST_PASSWORD", "admin@123")
	viper.SetDefault("PDF_SERVICE_URL", "http://localhost:8889")

	if err := viper.ReadInConfig(); err != nil {
		logger.Log.Sugar().Warnf("Aviso: arquivo .env não encontrado, usando variáveis de ambiente: %v", err)
	}

	if err := viperUnmarshal(&AppConfig); err != nil {
		logFatalf("Falha ao parsear configurações: %v", err)
	}
}

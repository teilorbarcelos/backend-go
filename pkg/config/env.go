package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Environment     string `mapstructure:"ENVIRONMENT"`
	Port            string `mapstructure:"PORT"`
	Host            string `mapstructure:"HOST"`
	DBUrl           string `mapstructure:"DATABASE_URL"`
	RedisUrl        string `mapstructure:"REDIS_URL"`
	RabbitMQUrl     string `mapstructure:"RABBITMQ_URL"`
	JWTSecret         string `mapstructure:"JWT_SECRET"`
	RateLimitMax      int    `mapstructure:"RATE_LIMIT_MAX"`
	RateLimitWindow   string `mapstructure:"RATE_LIMIT_WINDOW"`
	FirstUserEmail    string `mapstructure:"FIRST_USER"`
	FirstUserPassword string `mapstructure:"FIRST_PASSWORD"`
}

var AppConfig Config

func LoadConfig() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Default values
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("PORT", "3000")
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("RATE_LIMIT_MAX", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW", "1m")
	viper.SetDefault("FIRST_USER", "admin@email.com")
	viper.SetDefault("FIRST_PASSWORD", "admin@123")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Aviso: arquivo .env não encontrado, usando variáveis de ambiente: %v", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Falha ao parsear configurações: %v", err)
	}
}

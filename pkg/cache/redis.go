package cache

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/teilorbarcelos/backend-go/pkg/config"
)

var RedisClient *redis.Client

func ConnectRedis() {
	opts, err := redis.ParseURL(config.AppConfig.RedisUrl)
	if err != nil {
		log.Fatalf("Falha ao parsear a URL do Redis: %v", err)
	}

	RedisClient = redis.NewClient(opts)

	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Falha ao conectar no Redis: %v", err)
	}

	log.Println("Conexão com Redis estabelecida com sucesso.")
}

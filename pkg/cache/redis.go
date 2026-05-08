package cache

import (
	"context"
	"log"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/teilorbarcelos/backend-go/pkg/config"
)

var RedisClient *redis.Client

func ConnectRedis() {
	if config.AppConfig.Environment == "test" {
		mr, err := miniredis.Run()
		if err != nil {
			log.Fatalf("Falha ao iniciar miniredis: %v", err)
		}
		RedisClient = redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})
	} else {
		opts, err := redis.ParseURL(config.AppConfig.RedisUrl)
		if err != nil {
			log.Fatalf("Falha ao parsear a URL do Redis: %v", err)
		}
		RedisClient = redis.NewClient(opts)
	}

	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Falha ao conectar no Redis: %v", err)
	}

	log.Println("Conexão com Redis estabelecida com sucesso.")
}

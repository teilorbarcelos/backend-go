package auth

import (
	"os"
	"testing"

	"backend-go/pkg/cache"
	"backend-go/pkg/config"
	"backend-go/pkg/database"
)

func TestMain(m *testing.M) {
	os.Setenv("ENVIRONMENT", "test")
	config.LoadConfig()
	database.ConnectDB()
	cache.ConnectRedis()

	code := m.Run()
	os.Exit(code)
}

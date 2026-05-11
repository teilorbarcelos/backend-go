package repository

import (
	"os"
	"testing"

	"backend-go/pkg/config"
	"backend-go/pkg/database"
)

func TestMain(m *testing.M) {
	os.Setenv("ENVIRONMENT", "test")
	config.LoadConfig()
	database.ConnectDB()

	code := m.Run()
	os.Exit(code)
}

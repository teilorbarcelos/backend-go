package database

import (
	"log"

	"github.com/teilorbarcelos/backend-go/internal/core/audit"
	"github.com/teilorbarcelos/backend-go/internal/core/models"
	"github.com/teilorbarcelos/backend-go/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	if config.AppConfig.Environment == "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	if config.AppConfig.Environment == "test" {
		DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), gormConfig)
	} else {
		DB, err = gorm.Open(postgres.Open(config.AppConfig.DBUrl), gormConfig)
	}

	if err != nil {
		log.Fatalf("Falha ao conectar no banco de dados: %v", err)
	}

	// Registrar Hooks de Auditoria
	audit.RegisterAuditHooks(DB)

	// Rodar Migrations Automáticas
	log.Println("Rodando AutoMigrate...")
	err = DB.AutoMigrate(
		&models.AuditLog{},
		&models.Role{},
		&models.Feature{},
		&models.RoleFeature{},
		&models.Auth{},
		&models.User{},
		&models.Product{},
	)
	if err != nil {
		log.Fatalf("Erro no AutoMigrate: %v", err)
	}

	// Popular o banco com papéis padrão
	RunSeed(DB)

	log.Println("Conexão com PostgreSQL estabelecida com sucesso.")
}

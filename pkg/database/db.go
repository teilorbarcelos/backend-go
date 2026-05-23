package database

import (
	"time"

	"backend-go/pkg/logger"

	"backend-go/internal/core/models"
	"backend-go/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var DB *gorm.DB

var (
	logFatalf     = logger.Fatalf
	gormOpen      = gorm.Open
	dbAutoMigrate = func(db *gorm.DB, dst ...interface{}) error { return db.AutoMigrate(dst...) }
	runMigrations = defaultRunMigrations
	migrateNew    = func(sourceURL, databaseURL string) (migrator, error) {
		return migrate.New(sourceURL, databaseURL)
	}
)

type migrator interface {
	Up() error
}

func ConnectDB() {
	var err error

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	if config.AppConfig.Environment == "production" {
		gormConfig.Logger = gormlogger.Default.LogMode(gormlogger.Error)
	}

	DB, err = gormOpen(postgres.Open(config.AppConfig.DBUrl), gormConfig)

	if err != nil {
		logFatalf("Falha ao conectar no banco de dados: %v", err)
	}

	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
	}

	if config.AppConfig.Environment == "production" {
		runMigrations()
	} else {
		logger.Info("Rodando AutoMigrate...")
		DB.Exec("CREATE SCHEMA IF NOT EXISTS audit")
		err = dbAutoMigrate(
			DB,
			&models.AuditLog{},
			&models.ErrorLog{},
			&models.Role{},
			&models.Feature{},
			&models.RoleFeature{},
			&models.Auth{},
			&models.User{},
			&models.Product{},
		)
		if err != nil {
			logFatalf("Erro no AutoMigrate: %v", err)
		}
	}

	RunSeed(DB)

	logger.Info("Conexão com PostgreSQL estabelecida com sucesso.")
}

func defaultRunMigrations() {
	m, err := migrateNew(
		"file://database/migrations",
		config.AppConfig.DBUrl,
	)
	if err != nil {
		logFatalf("Falha ao preparar migrações: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logFatalf("Falha ao executar migrações: %v", err)
	}

	logger.Info("Migrações aplicadas com sucesso.")
}

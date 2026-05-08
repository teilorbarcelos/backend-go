package database

import (
	"errors"
	"testing"

	"backend-go/internal/core/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestRunSeed_HashError(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	
	// Migrate tables needed for RunSeed to execute without errors before reaching the hash check
	db.AutoMigrate(
		&models.AuditLog{},
		&models.Role{},
		&models.Feature{},
		&models.RoleFeature{},
		&models.Auth{},
		&models.User{},
	)
	
	origHash := hashPassword
	defer func() { hashPassword = origHash }()
	
	// Force HashPassword to fail
	hashPassword = func(password string) (string, error) {
		return "", errors.New("mock hashing error")
	}
	
	// RunSeed should log the error and continue
	RunSeed(db)
}

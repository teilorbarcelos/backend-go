package database

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"backend-go/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConnectDB(t *testing.T) {
	// Backup original values
	origEnv := config.AppConfig.Environment
	origDBUrl := config.AppConfig.DBUrl
	origFatalf := logFatalf
	origGormOpen := gormOpen
	origAutoMigrate := dbAutoMigrate
	origDB := DB

	defer func() {
		config.AppConfig.Environment = origEnv
		config.AppConfig.DBUrl = origDBUrl
		logFatalf = origFatalf
		gormOpen = origGormOpen
		dbAutoMigrate = origAutoMigrate
		DB = origDB
	}()

	t.Run("Success in test mode", func(t *testing.T) {
		config.AppConfig.Environment = "test"
		gormOpen = gorm.Open
		dbAutoMigrate = func(db *gorm.DB, dst ...interface{}) error { return nil }
		logFatalf = origFatalf
		
		ConnectDB()
		assert.NotNil(t, DB)
	})

	t.Run("Success in production mode", func(t *testing.T) {
		config.AppConfig.Environment = "production"
		// Mock gormOpen to use sqlite even in "production" mode for the test
		gormOpen = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
			return gorm.Open(sqlite.Open("file::memory:?cache=shared"), opts...)
		}
		dbAutoMigrate = func(db *gorm.DB, dst ...interface{}) error { return nil }
		logFatalf = origFatalf

		ConnectDB()
		assert.NotNil(t, DB)
		// Verification that we passed through the production check could be logger level check, 
		// but since we mocked gormOpen we can't easily check the internal gormConfig.
		// However, the branch is covered.
	})

	t.Run("Failure on gorm.Open", func(t *testing.T) {
		config.AppConfig.Environment = "test"
		gormOpen = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
			return nil, errors.New("connection error")
		}
		
		logFatalf = func(format string, v ...interface{}) {
			panic("fatal: connection")
		}

		assert.PanicsWithValue(t, "fatal: connection", func() {
			ConnectDB()
		})
	})

	t.Run("Failure on AutoMigrate", func(t *testing.T) {
		config.AppConfig.Environment = "test"
		gormOpen = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
			return gorm.Open(sqlite.Open("file::memory:?cache=shared"), opts...)
		}
		dbAutoMigrate = func(db *gorm.DB, dst ...interface{}) error {
			return errors.New("migration error")
		}
		
		logFatalf = func(format string, v ...interface{}) {
			panic("fatal: migration")
		}

		assert.PanicsWithValue(t, "fatal: migration", func() {
			ConnectDB()
		})
	})
}

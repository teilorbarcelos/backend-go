package auth

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"backend-go/internal/core/models"
	"backend-go/pkg/cache"
	"backend-go/pkg/config"
	"backend-go/pkg/database"
	"backend-go/pkg/security"
	"github.com/redis/go-redis/v9"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestMain(m *testing.M) {
	os.Setenv("ENVIRONMENT", "test")
	config.LoadConfig()
	database.ConnectDB()
	cache.ConnectRedis()

	code := m.Run()
	os.Exit(code)
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewAuthService(mockRepo)

	password := "password123"
	hashedPassword, _ := security.HashPassword(password)

	user := &models.User{
		BaseModel: models.BaseModel{ID: "1"},
		Email:     "test@test.com",
		Active:    true,
		Auth: &models.Auth{
			Password: &hashedPassword,
		},
		Role: models.Role{
			BaseModel: models.BaseModel{ID: "admin"},
			Name:      "Admin",
			RoleFeature: []models.RoleFeature{
				{IDFeature: "f1", View: true},
			},
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("FindByEmail", "test@test.com").Return(user, nil)
		res, err := service.Login("test@test.com", password)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, res.Valid)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo.On("FindByEmail", "notfound@test.com").Return(nil, os.ErrNotExist)
		res, err := service.Login("notfound@test.com", password)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "usuário não encontrado", err.Error())
	})

	t.Run("Inactive User", func(t *testing.T) {
		inactiveUser := *user
		inactiveUser.Active = false
		mockRepo.On("FindByEmail", "inactive@test.com").Return(&inactiveUser, nil)
		_, err := service.Login("inactive@test.com", password)
		assert.Error(t, err)
		assert.Equal(t, "conta desativada ou removida", err.Error())
	})

	t.Run("No Auth Configured", func(t *testing.T) {
		noAuthUser := *user
		noAuthUser.Auth = nil
		mockRepo.On("FindByEmail", "noauth@test.com").Return(&noAuthUser, nil)
		_, err := service.Login("noauth@test.com", password)
		assert.Error(t, err)
		assert.Equal(t, "autenticação não configurada para este usuário", err.Error())
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mockRepo.On("FindByEmail", "test@test.com").Return(user, nil)
		_, err := service.Login("test@test.com", "wrong")
		assert.Error(t, err)
		assert.Equal(t, "senha inválida", err.Error())
	})

	t.Run("Redis Error", func(t *testing.T) {
		oldClient := cache.RedisClient
		cache.RedisClient = redis.NewClient(&redis.Options{Addr: "localhost:1"})
		defer func() { cache.RedisClient = oldClient }()

		mockRepo.On("FindByEmail", "test@test.com").Return(user, nil)
		res, err := service.Login("test@test.com", password)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Token Error", func(t *testing.T) {
		service.GenerateToken = func(id, email, idRole string, perms []security.Permission) (string, error) {
			return "", errors.New("token err")
		}
		defer func() { service.GenerateToken = security.GenerateToken }()

		mockRepo.On("FindByEmail", "test@test.com").Return(user, nil)
		_, err := service.Login("test@test.com", password)
		assert.Error(t, err)
		assert.Equal(t, "token err", err.Error())
	})
}

func TestAuthService_GetMe(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewAuthService(mockRepo)

	user := &models.User{
		BaseModel: models.BaseModel{ID: "1"},
		Email:     "test@test.com",
		Active:    true,
		Role: models.Role{
			BaseModel: models.BaseModel{ID: "admin"},
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("FindByEmail", "test@test.com").Return(user, nil)
		res, err := service.GetMe("test@test.com")
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("FindByEmail", "error@test.com").Return(nil, os.ErrNotExist)
		_, err := service.GetMe("error@test.com")
		assert.Error(t, err)
	})

	t.Run("Inactive", func(t *testing.T) {
		inactive := *user
		inactive.Active = false
		mockRepo.On("FindByEmail", "inactive@test.com").Return(&inactive, nil)
		_, err := service.GetMe("inactive@test.com")
		assert.Error(t, err)
	})

	t.Run("IsDeleted", func(t *testing.T) {
		deleted := *user
		deleted.IsDeleted = true
		mockRepo.On("FindByEmail", "deleted@test.com").Return(&deleted, nil)
		_, err := service.GetMe("deleted@test.com")
		assert.Error(t, err)
	})

	t.Run("Token Error", func(t *testing.T) {
		service.GenerateToken = func(id, email, idRole string, perms []security.Permission) (string, error) {
			return "", errors.New("token err")
		}
		defer func() { service.GenerateToken = security.GenerateToken }()

		mockRepo.On("FindByEmail", "test@test.com").Return(user, nil)
		_, err := service.GetMe("test@test.com")
		assert.Error(t, err)
		assert.Equal(t, "token err", err.Error())
	})
}

package user

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"backend-go/internal/core/models"
	"backend-go/internal/infra/session"
	"backend-go/pkg/config"
	"backend-go/pkg/database"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(id string, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id string, preloads ...string) (*models.User, error) {
	// Variadic arguments in mock require special handling if we want to match them exactly,
	// but usually we can just pass them to Called.
	args := m.Called(id, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string, preloads ...string) (*models.User, error) {
	args := m.Called(email, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(authID string, password string) error {
	args := m.Called(authID, password)
	return args.Error(0)
}


func (m *MockUserRepository) SearchPaginated(params database.FilterParams, filterable map[string]database.FilterConfig, searchable []database.SearchConfig, preloads ...string) ([]models.User, int64, error) {
	args := m.Called(params, filterable, searchable, preloads)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) WithContext(ctx context.Context) UserRepositoryI {
	// Usually WithContext returns itself or a new mock, but for simplicity we return the same mock
	return m
}


func TestUserService_Create(t *testing.T) {
	repo := NewUserRepository(database.DB)
	sessionMgr := session.NewSessionManager()
	service := NewUserService(repo, sessionMgr)

	dto := CreateUserDTO{
		Name:     "Test User",
		Email:    "test-create@example.com",
		Password: "password123",
		IDRole:   "administrator",
	}

	ctx := context.Background()
	user, err := service.Create(ctx, dto)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, dto.Name, user.Name)
	assert.Equal(t, dto.Email, user.Email)
	assert.NotEmpty(t, user.ID)
	assert.NotNil(t, user.Auth)
}

func TestUserService_Update(t *testing.T) {
	repo := NewUserRepository(database.DB)
	sessionMgr := session.NewSessionManager()
	service := NewUserService(repo, sessionMgr)
	ctx := context.Background()

	// 1. Setup a regular user
	user, err := service.Create(ctx, CreateUserDTO{
		Name:     "Old Name",
		Email:    "old-update-unique@email.com",
		Password: "password",
		IDRole:   "administrator",
	})
	assert.NoError(t, err)

	t.Run("Update name and role", func(t *testing.T) {
		active := true
		updated, err := service.Update(ctx, user.ID, UpdateUserDTO{
			Name:   "New Name",
			IDRole: "manager",
			Active: &active,
		})
		if assert.NoError(t, err) && assert.NotNil(t, updated) {
			assert.Equal(t, "New Name", updated.Name)
			assert.Equal(t, "manager", updated.IDRole)
		}
	})

	t.Run("Update password", func(t *testing.T) {
		_, err := service.Update(ctx, user.ID, UpdateUserDTO{
			Password: "new-password",
		})
		assert.NoError(t, err)
	})

	t.Run("Update email", func(t *testing.T) {
		updated, err := service.Update(ctx, user.ID, UpdateUserDTO{
			Email: "new-email-unique@email.com",
		})
		if assert.NoError(t, err) && assert.NotNil(t, updated) {
			assert.Equal(t, "new-email-unique@email.com", updated.Email)
		}
	})

	t.Run("Admin protections", func(t *testing.T) {
		// Ensure admin user exists
		adminEmail := config.AppConfig.FirstUserEmail
		// Try to find it first
		foundAdmin, err := repo.WithContext(ctx).FindByEmail(adminEmail, "Auth")
		var admin *models.User
		if err != nil {
			// Create it if not found
			newAdmin, createErr := service.Create(ctx, CreateUserDTO{
				Name:     "Admin",
				Email:    adminEmail,
				Password: "password",
				IDRole:   "administrator",
			})
			assert.NoError(t, createErr, "should be able to create admin user")
			admin = newAdmin
		} else {
			admin = foundAdmin
		}
		
		assert.NotEmpty(t, admin.ID, "admin user ID should not be empty")
		assert.Equal(t, adminEmail, admin.Email, "admin email should match config")

		// Try to deactivate admin
		active := false
		_, err = service.Update(ctx, admin.ID, UpdateUserDTO{
			Active: &active,
		})
		if assert.Error(t, err) {
			assert.Equal(t, "o usuário administrador inicial não pode ser desativado", err.Error())
		}

		// Try to change admin email
		_, err = service.Update(ctx, admin.ID, UpdateUserDTO{
			Email: "other@email.com",
		})
		if assert.Error(t, err) {
			assert.Equal(t, "o email do usuário administrador inicial não pode ser alterado", err.Error())
		}
		
		// Change admin name (should be allowed)
		updated, err := service.Update(ctx, admin.ID, UpdateUserDTO{
			Name: "Updated Admin",
		})
		assert.NoError(t, err)
		assert.Equal(t, "Updated Admin", updated.Name)
	})
}

func TestUserService_List(t *testing.T) {
	repo := NewUserRepository(database.DB)
	sessionMgr := session.NewSessionManager()
	service := NewUserService(repo, sessionMgr)

	// Create a user first
	ctx := context.Background()
	_, err := service.Create(ctx, CreateUserDTO{
		Name:     "List Test",
		Email:    "list-unique-test@example.com",
		Password: "password123",
		IDRole:   "administrator",
	})
	assert.NoError(t, err)

	params := database.FilterParams{
		Pagination: database.Pagination{
			Page:  1,
			Limit: 10,
		},
		Filters: map[string]interface{}{},
	}

	users, total, err := service.List(ctx, params)
	if assert.NoError(t, err) {
		assert.True(t, total > 0)
		assert.NotEmpty(t, users)
	}
}

func TestUserService_Delete(t *testing.T) {
	repo := NewUserRepository(database.DB)
	sessionMgr := session.NewSessionManager()
	service := NewUserService(repo, sessionMgr)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		user, err := service.Create(ctx, CreateUserDTO{
			Name:     "Delete Me",
			Email:    "delete-unique-service@me.com",
			Password: "password",
			IDRole:   "administrator",
		})
		if assert.NoError(t, err) && assert.NotNil(t, user) {
			err = service.Delete(ctx, user.ID)
			assert.NoError(t, err)
	
			// Verify it's gone
			_, err = service.GetByID(ctx, user.ID)
			assert.Error(t, err)
		}
	})

	t.Run("Admin protection", func(t *testing.T) {
		// Admin is created in Update test or already exists
		// We can just find it by email
		if u, err := repo.WithContext(ctx).FindByEmail(config.AppConfig.FirstUserEmail); err == nil {
			err = service.Delete(ctx, u.ID)
			assert.Error(t, err)
			assert.Equal(t, "o usuário administrador inicial não pode ser excluído", err.Error())
		}
	})
}

func TestUserService_SetStatus(t *testing.T) {
	repo := NewUserRepository(database.DB)
	sessionMgr := session.NewSessionManager()
	service := NewUserService(repo, sessionMgr)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		user, err := service.Create(ctx, CreateUserDTO{
			Name:     "Status Test",
			Email:    "status-unique-service@email.com",
			Password: "password",
			IDRole:   "administrator",
		})
		if assert.NoError(t, err) && assert.NotNil(t, user) {
			err = service.SetStatus(ctx, user.ID, false)
			assert.NoError(t, err)
	
			updated, _ := service.GetByID(ctx, user.ID)
			if assert.NotNil(t, updated) {
				assert.False(t, updated.Active)
			}
		}
	})

	t.Run("Admin protection", func(t *testing.T) {
		if u, err := repo.WithContext(ctx).FindByEmail(config.AppConfig.FirstUserEmail); err == nil {
			err = service.SetStatus(ctx, u.ID, false)
			assert.Error(t, err)
			assert.Equal(t, "o usuário administrador inicial não pode ser desativado", err.Error())
			
			// Activating admin should be allowed (even if already active)
			err = service.SetStatus(ctx, u.ID, true)
			assert.NoError(t, err)
		}
	})
}

func TestUserService_ErrorPaths(t *testing.T) {
	sessionMgr := session.NewSessionManager()
	ctx := context.Background()

	t.Run("Create Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		mockRepo.On("Create", mock.Anything).Return(errors.New("db error")).Once()
		_, err := service.Create(ctx, CreateUserDTO{Password: "pass"})
		assert.Error(t, err)
	})

	t.Run("Update FindByID Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		mockRepo.On("FindByID", "1", mock.Anything).Return(nil, errors.New("not found")).Once()
		_, err := service.Update(ctx, "1", UpdateUserDTO{})
		assert.Error(t, err)
	})

	t.Run("Update Repo Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		user := &models.User{Email: "test@test.com"}
		mockRepo.On("FindByID", "1", mock.Anything).Return(user, nil).Once()
		mockRepo.On("Update", "1", mock.Anything).Return(errors.New("update error")).Once()
		_, err := service.Update(ctx, "1", UpdateUserDTO{Name: "New"})
		assert.Error(t, err)
	})

	t.Run("Update Password Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		idAuth := "auth-id"
		user := &models.User{Email: "test@test.com", IDAuth: &idAuth}
		mockRepo.On("FindByID", "1", mock.Anything).Return(user, nil).Once() // Only once because it fails early
		mockRepo.On("UpdatePassword", idAuth, mock.Anything).Return(errors.New("pass error")).Once()
		_, err := service.Update(ctx, "1", UpdateUserDTO{Password: "new-pass"})
		assert.Error(t, err)
	})

	t.Run("Create Password Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		// Bcrypt has a maximum password length (72 bytes). 
		// Providing a very long password should trigger an error in HashPassword.
		longPass := make([]byte, 100)
		for i := range longPass {
			longPass[i] = 'a'
		}
		_, err := service.Create(ctx, CreateUserDTO{Password: string(longPass)})
		assert.Error(t, err)
	})

	t.Run("Delete FindByID Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		mockRepo.On("FindByID", "1", mock.Anything).Return(nil, errors.New("not found")).Once()
		err := service.Delete(ctx, "1")
		assert.Error(t, err)
	})

	t.Run("Delete Repo Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		user := &models.User{Email: "test@test.com"}
		mockRepo.On("FindByID", "1", mock.Anything).Return(user, nil).Once()
		mockRepo.On("Update", "1", mock.Anything).Return(nil).Once()
		mockRepo.On("Delete", "1").Return(errors.New("delete error")).Once()
		err := service.Delete(ctx, "1")
		assert.Error(t, err)
	})

	t.Run("Delete Anonymize/Update Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		user := &models.User{Email: "test@test.com"}
		mockRepo.On("FindByID", "1", mock.Anything).Return(user, nil).Once()
		mockRepo.On("Update", "1", mock.Anything).Return(errors.New("update error")).Once()
		err := service.Delete(ctx, "1")
		assert.Error(t, err)
		assert.Equal(t, "update error", err.Error())
	})

	t.Run("SetStatus FindByID Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		mockRepo.On("FindByID", "1", mock.Anything).Return(nil, errors.New("not found")).Once()
		err := service.SetStatus(ctx, "1", true)
		assert.Error(t, err)
	})

	t.Run("SetStatus Repo Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, sessionMgr)
		user := &models.User{Email: "test@test.com"}
		mockRepo.On("FindByID", "1", mock.Anything).Return(user, nil).Once()
		mockRepo.On("Update", "1", mock.Anything).Return(errors.New("update error")).Once()
		err := service.SetStatus(ctx, "1", true)
		assert.Error(t, err)
	})
}

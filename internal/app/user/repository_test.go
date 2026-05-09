package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"backend-go/internal/core/models"
	"backend-go/pkg/database"
)

func TestUserRepository_FindAllWithRole(t *testing.T) {
	var repo UserRepositoryI = NewUserRepository(database.DB)
	ctx := context.Background()
	repo = repo.WithContext(ctx)

	// Create a role and some users
	role := models.Role{
		Name:        "Repo Test Role",
		Description: "Role for repo testing",
	}
	database.DB.Create(&role)

	user1 := models.User{
		Name:   "Repo User 1",
		Email:  "repo1@test.com",
		IDRole: role.ID,
	}
	user2 := models.User{
		Name:   "Repo User 2",
		Email:  "repo2@test.com",
		IDRole: role.ID,
	}
	database.DB.Create(&user1)
	database.DB.Create(&user2)

	t.Run("Success without filters", func(t *testing.T) {
		users, total, err := repo.FindAllWithRole(nil, 0, 10)
		assert.NoError(t, err)
		assert.True(t, total >= 2)
		assert.NotEmpty(t, users)
		
		// Verify role is preloaded
		found := false
		for _, u := range users {
			if u.ID == user1.ID {
				assert.Equal(t, role.ID, u.Role.ID)
				assert.Equal(t, role.Name, u.Role.Name)
				found = true
			}
		}
		assert.True(t, found)
	})

	t.Run("Success with filters", func(t *testing.T) {
		// Note: we use "name = ?" because of how FindAllWithRole is implemented
		filter := map[string]interface{}{"name = ?": "Repo User 1"}
		users, total, err := repo.FindAllWithRole(filter, 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, users, 1)
		assert.Equal(t, "Repo User 1", users[0].Name)
	})

	t.Run("With offset and limit", func(t *testing.T) {
		users, total, err := repo.FindAllWithRole(nil, 1, 1)
		assert.NoError(t, err)
		assert.True(t, total >= 2)
		assert.Len(t, users, 1)
	})

	t.Run("DB Error", func(t *testing.T) {
		// Providing an invalid column in filter should trigger a DB error
		filter := map[string]interface{}{"invalid_column = ?": "value"}
		_, _, err := repo.FindAllWithRole(filter, 0, 10)
		assert.Error(t, err)
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	repo := NewUserRepository(database.DB)
	ctx := context.Background()

	// Create a user
	user := models.User{
		Name:  "FindByEmail User",
		Email: "findbyemail@test.com",
	}
	database.DB.Create(&user)

	t.Run("Success", func(t *testing.T) {
		found, err := repo.WithContext(ctx).FindByEmail(user.Email)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		found, err := repo.WithContext(ctx).FindByEmail("nonexistent@test.com")
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"backend-go/internal/core/models"
	"backend-go/pkg/database"
)

func TestUserRepository_SearchPaginated(t *testing.T) {
	var repo UserRepositoryI = NewUserRepository(database.DB)
	ctx := context.Background()
	repo = repo.WithContext(ctx)

	// Create a role and some users
	role := models.Role{
		Name:        "Repo Search Role",
		Description: "Role for repo testing",
	}
	database.DB.Create(&role)

	user1 := models.User{
		Name:   "Search User 1",
		Email:  "search1@test.com",
		IDRole: role.ID,
	}
	user2 := models.User{
		Name:   "Search User 2",
		Email:  "search2@test.com",
		IDRole: role.ID,
	}
	database.DB.Create(&user1)
	database.DB.Create(&user2)

	filterable := map[string]database.FilterConfig{
		"name": {Operator: "contains"},
	}

	t.Run("Success without filters", func(t *testing.T) {
		params := database.FilterParams{
			Pagination: database.Pagination{Limit: 10},
		}
		users, total, err := repo.SearchPaginated(params, filterable, nil, "Role")
		assert.NoError(t, err)
		assert.True(t, total >= 2)
		assert.NotEmpty(t, users)
		
		found := false
		for _, u := range users {
			if u.ID == user1.ID {
				assert.Equal(t, role.ID, u.Role.ID)
				found = true
			}
		}
		assert.True(t, found)
	})

	t.Run("Success with filters", func(t *testing.T) {
		params := database.FilterParams{
			Pagination: database.Pagination{Limit: 10},
			Filters: map[string]interface{}{"name": "Search User 1"},
		}
		users, total, err := repo.SearchPaginated(params, filterable, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, users, 1)
		assert.Equal(t, "Search User 1", users[0].Name)
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

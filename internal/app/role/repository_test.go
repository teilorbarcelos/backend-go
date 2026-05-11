package role

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"backend-go/internal/core/models"
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

func TestRoleRepository_Create(t *testing.T) {
	repo := NewRoleRepository(database.DB)
	role := &models.Role{
		Name:        "Test Role",
		Description: "Test Description",
		Active:      true,
	}

	err := repo.Create(role)
	assert.NoError(t, err)
	assert.NotEmpty(t, role.ID)
}

func TestRoleRepository_FindByID(t *testing.T) {
	repo := NewRoleRepository(database.DB)
	role := &models.Role{
		Name:        "Find Test",
		Description: "Find Description",
	}
	repo.Create(role)

	found, err := repo.FindByID(role.ID)
	assert.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)
}

func TestRoleRepository_CreateWithPermissions(t *testing.T) {
	repo := NewRoleRepository(database.DB)
	
	t.Run("Success", func(t *testing.T) {
		role := &models.Role{Name: "With Perms", Description: "Desc"}
		perms := []models.RoleFeature{
			{IDFeature: "feat1", View: true},
		}
		err := repo.CreateWithPermissions(role, perms)
		assert.NoError(t, err)
		assert.NotEmpty(t, role.ID)
	})

	t.Run("Error - ID Collision", func(t *testing.T) {
		role1 := &models.Role{Name: "R1", Description: "D"}
		role1.ID = "fixed-id"
		repo.Create(role1)

		role2 := &models.Role{Name: "R2", Description: "D"}
		role2.ID = "fixed-id"
		err := repo.CreateWithPermissions(role2, nil)
		assert.Error(t, err)
	})

	t.Run("Error - Permission Violation", func(t *testing.T) {
		role := &models.Role{Name: "Perm Error", Description: "D"}
		perms := []models.RoleFeature{
			{IDFeature: "feat_same", View: true},
			{IDFeature: "feat_same", View: true}, // Duplicate PK
		}
		err := repo.CreateWithPermissions(role, perms)
		assert.Error(t, err)
	})
}

func TestRoleRepository_UpdateWithPermissions(t *testing.T) {
	repo := NewRoleRepository(database.DB)
	role := &models.Role{Name: "To Update", Description: "D"}
	repo.Create(role)

	t.Run("Success", func(t *testing.T) {
		role.Name = "Updated Name"
		perms := []models.RoleFeature{
			{IDFeature: "feat2", View: true},
		}
		err := repo.UpdateWithPermissions(role.ID, role, perms)
		assert.NoError(t, err)
	})

	t.Run("Error - Permission Violation", func(t *testing.T) {
		perms := []models.RoleFeature{
			{IDFeature: "f1", View: true},
			{IDFeature: "f1", View: true}, // Duplicate
		}
		err := repo.UpdateWithPermissions(role.ID, role, perms)
		assert.Error(t, err)
	})

	t.Run("Error - Update Violation", func(t *testing.T) {
		database.DB.Exec("CREATE TRIGGER fail_update BEFORE UPDATE ON role BEGIN SELECT RAISE(ABORT, 'forced failure'); END;")
		defer database.DB.Exec("DROP TRIGGER fail_update;")

		err := repo.UpdateWithPermissions(role.ID, role, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "forced failure")
	})

	t.Run("Error - Delete Violation", func(t *testing.T) {
		database.DB.Exec("CREATE TRIGGER fail_delete BEFORE DELETE ON role_feature BEGIN SELECT RAISE(ABORT, 'forced failure'); END;")
		defer database.DB.Exec("DROP TRIGGER fail_delete;")

		err := repo.UpdateWithPermissions(role.ID, role, []models.RoleFeature{{IDFeature: "f1"}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "forced failure")
	})

	t.Run("Success - Empty Permissions", func(t *testing.T) {
		err := repo.UpdateWithPermissions(role.ID, role, []models.RoleFeature{})
		assert.NoError(t, err)
	})
}

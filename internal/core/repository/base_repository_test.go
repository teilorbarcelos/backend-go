package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teilorbarcelos/backend-go/internal/core/models"
	"github.com/teilorbarcelos/backend-go/pkg/database"
)

func TestBaseRepository_FindAll(t *testing.T) {
	repo := NewBaseRepository[models.Product](database.DB)

	// Setup
	p1 := models.Product{Name: "Product 1", SKU: "SKU1", Category: "Cat1", Price: 10.0}
	p2 := models.Product{Name: "Product 2", SKU: "SKU2", Category: "Cat1", Price: 20.0}
	database.DB.Create(&p1)
	database.DB.Create(&p2)

	t.Run("Find All without filter", func(t *testing.T) {
		products, total, err := repo.FindAll(nil, 0, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.NotEmpty(t, products)
	})

	t.Run("Find All with filter", func(t *testing.T) {
		filter := map[string]interface{}{"name": "Product 1"}
		products, total, err := repo.FindAll(filter, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Product 1", products[0].Name)
	})

	t.Run("Find All with pagination", func(t *testing.T) {
		products, total, err := repo.FindAll(nil, 1, 1)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.Len(t, products, 1)
	})

	t.Run("Find All error case", func(t *testing.T) {
		filter := map[string]interface{}{"non_existent_column": "value"}
		_, _, err := repo.FindAll(filter, 0, 0)
		assert.Error(t, err)
	})
}

func TestBaseRepository_FindByID(t *testing.T) {
	repo := NewBaseRepository[models.Product](database.DB)

	p := models.Product{Name: "FindByID Test", SKU: "SKUID", Category: "Cat", Price: 10.0}
	database.DB.Create(&p)

	t.Run("Success", func(t *testing.T) {
		found, err := repo.FindByID(p.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, p.ID, found.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		found, err := repo.FindByID("non-existent-id")
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

func TestBaseRepository_HardDelete(t *testing.T) {
	repo := NewBaseRepository[models.Product](database.DB)

	p := models.Product{Name: "HardDelete Test", SKU: "SKUHD", Category: "Cat", Price: 10.0}
	database.DB.Create(&p)

	err := repo.HardDelete(p.ID)
	assert.NoError(t, err)

	// Verify it's gone even from unscoped
	var found models.Product
	err = database.DB.Unscoped().Where("id = ?", p.ID).First(&found).Error
	assert.Error(t, err)
}

func TestBaseRepository_SearchPaginated_Coverage(t *testing.T) {
	repo := NewBaseRepository[models.Product](database.DB)

	t.Run("Success with ignoreDefaultFilters", func(t *testing.T) {
		params := database.FilterParams{
			Filters: map[string]interface{}{
				"ignoreDefaultFilters": true,
			},
		}
		_, _, err := repo.SearchPaginated(params, nil)
		assert.NoError(t, err)
	})

	t.Run("Success with preloads", func(t *testing.T) {
		// User has Role and Auth preloads
		userRepo := NewBaseRepository[models.User](database.DB)
		params := database.FilterParams{}
		_, _, err := userRepo.SearchPaginated(params, nil, "Role")
		assert.NoError(t, err)
	})

	t.Run("Error Case - Count", func(t *testing.T) {
		// Using an invalid filter to trigger a SQL error during Count
		params := database.FilterParams{
			Filters: map[string]interface{}{
				"non_existent_column": "value",
			},
		}
		_, _, err := repo.SearchPaginated(params, nil)
		assert.Error(t, err)
	})
}

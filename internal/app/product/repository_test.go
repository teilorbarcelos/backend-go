package product

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"backend-go/internal/core/models"
	"backend-go/pkg/config"
	"backend-go/pkg/database"
)

func TestMain(m *testing.M) {
	os.Setenv("ENVIRONMENT", "test")
	config.LoadConfig()
	database.ConnectDB()

	code := m.Run()
	os.Exit(code)
}

func TestProductRepository_Create(t *testing.T) {
	repo := NewProductRepository(database.DB)
	product := &models.Product{
		Name:     "Test Product",
		SKU:      "SKU-123",
		Category: "Test",
		Price:    10.5,
		Active:   true,
	}

	err := repo.Create(product)
	assert.NoError(t, err)
	assert.NotEmpty(t, product.ID)
}

func TestProductRepository_FindByID(t *testing.T) {
	repo := NewProductRepository(database.DB)
	product := &models.Product{
		Name:     "Find Test",
		SKU:      "SKU-FIND",
		Category: "Test",
		Price:    20.0,
	}
	repo.Create(product)

	found, err := repo.FindByID(product.ID)
	assert.NoError(t, err)
	assert.Equal(t, product.ID, found.ID)
}

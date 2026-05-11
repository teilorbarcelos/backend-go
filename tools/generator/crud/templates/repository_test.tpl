package {{.LowerName}}

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func Test{{.Name}}Repository_Create(t *testing.T) {
	repo := New{{.Name}}Repository(database.DB)
	entity := &{{.Name}}{
		Name: "Test {{.Name}}",
	}

	err := repo.Create(entity)
	assert.NoError(t, err)
	assert.NotEmpty(t, entity.ID)
}

func Test{{.Name}}Repository_FindByID(t *testing.T) {
	repo := New{{.Name}}Repository(database.DB)
	entity := &{{.Name}}{
		Name: "Find Test",
	}
	repo.Create(entity)

	found, err := repo.FindByID(entity.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.ID, found.ID)
	assert.Equal(t, "Find Test", found.Name)
}

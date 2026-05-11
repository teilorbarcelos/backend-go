package {{.LowerName}}

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"backend-go/pkg/database"
)

func Test{{.Name}}Service_Create(t *testing.T) {
	repo := New{{.Name}}Repository(database.DB)
	service := New{{.Name}}Service(repo)
	ctx := context.Background()

	entity := &{{.Name}}{
		Name: "Service Test",
	}

	err := service.Create(ctx, entity)
	assert.NoError(t, err)
	assert.NotEmpty(t, entity.ID)
}

func Test{{.Name}}Service_List(t *testing.T) {
	repo := New{{.Name}}Repository(database.DB)
	service := New{{.Name}}Service(repo)
	ctx := context.Background()

	params := database.FilterParams{
		Pagination: database.Pagination{
			Page:  1,
			Limit: 10,
		},
	}

	items, total, err := service.List(ctx, params)
	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.True(t, total >= 0)
}

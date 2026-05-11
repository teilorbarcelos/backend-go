package {{.LowerName}}

import (
	"backend-go/internal/core/models"
	"backend-go/internal/core/repository"
	"gorm.io/gorm"
)

type {{.Name}} struct {
	models.BaseModel
	Name string `gorm:"type:varchar(255);not null" json:"name"`
}

type {{.Name}}Repository struct {
	*repository.BaseRepository[{{.Name}}]
}

func New{{.Name}}Repository(db *gorm.DB) *{{.Name}}Repository {
	return &{{.Name}}Repository{
		BaseRepository: repository.NewBaseRepository[{{.Name}}](db),
	}
}

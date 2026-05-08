package product

import (
	"github.com/teilorbarcelos/backend-go/internal/core/models"
	"github.com/teilorbarcelos/backend-go/internal/core/repository"
	"gorm.io/gorm"
)

type ProductRepository struct {
	repository.BaseRepository[models.Product]
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		BaseRepository: *repository.NewBaseRepository[models.Product](db),
	}
}

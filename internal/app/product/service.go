package product

import (
	"github.com/teilorbarcelos/backend-go/internal/core/models"
)

type ProductService struct {
	Repo *ProductRepository
}

func NewProductService(repo *ProductRepository) *ProductService {
	return &ProductService{Repo: repo}
}

type CreateProductDTO struct {
	Name        string  `json:"name" binding:"required"`
	SKU         string  `json:"sku" binding:"required"`
	Category    string  `json:"category" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Stock       int     `json:"stock"`
	Description string  `json:"description"`
}

func (s *ProductService) Create(dto CreateProductDTO) (*models.Product, error) {
	product := &models.Product{
		Name:        dto.Name,
		SKU:         dto.SKU,
		Category:    dto.Category,
		Price:       dto.Price,
		Stock:       dto.Stock,
		Description: dto.Description,
		Active:      true,
	}
	err := s.Repo.Create(product)
	return product, err
}

func (s *ProductService) Update(id string, updates map[string]interface{}) (*models.Product, error) {
	err := s.Repo.Update(id, updates)
	if err != nil {
		return nil, err
	}
	return s.Repo.FindByID(id)
}

func (s *ProductService) List(offset, limit int) ([]models.Product, int64, error) {
	return s.Repo.FindAll(nil, offset, limit)
}

func (s *ProductService) GetByID(id string) (*models.Product, error) {
	return s.Repo.FindByID(id)
}

func (s *ProductService) Delete(id string) error {
	return s.Repo.Delete(id)
}

func (s *ProductService) SetStatus(id string, active bool) error {
	return s.Repo.Update(id, map[string]interface{}{"active": active})
}

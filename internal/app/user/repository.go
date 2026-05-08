package user

import (
	"github.com/teilorbarcelos/backend-go/internal/core/models"
	"github.com/teilorbarcelos/backend-go/internal/core/repository"
	"gorm.io/gorm"
)

type UserRepository struct {
	repository.BaseRepository[models.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: *repository.NewBaseRepository[models.User](db),
	}
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Auth").Preload("Role").Where("id = ? AND is_deleted = ?", id, false).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAllWithRole(filter map[string]interface{}, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.DB.Model(&models.User{}).Preload("Role").Where("is_deleted = ?", false)

	for k, v := range filter {
		query = query.Where(k, v)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err = query.Find(&users).Error
	return users, total, err
}

package auth

import (
	"backend-go/internal/core/models"
	"backend-go/internal/core/repository"
	"gorm.io/gorm"
)

type AuthRepositoryI interface {
	FindByEmail(email string) (*models.User, error)
}

type AuthRepository struct {
	repository.BaseRepository[models.User]
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{
		BaseRepository: *repository.NewBaseRepository[models.User](db),
	}
}

func (r *AuthRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Auth").Preload("Role").Preload("Role.RoleFeature").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

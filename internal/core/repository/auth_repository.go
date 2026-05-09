package repository

import (
	"backend-go/internal/core/models"
	"gorm.io/gorm"
)

type AuthRepository struct {
	BaseRepository[models.User]
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{
		BaseRepository: *NewBaseRepository[models.User](db),
	}
}

// FindByEmail busca um usuário pelo email incluindo a entidade Auth para validação de senha.
func (r *AuthRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Auth").Preload("Role").Preload("Role.RoleFeature").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

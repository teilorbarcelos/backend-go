package role

import (
	"github.com/teilorbarcelos/backend-go/internal/core/models"
	"github.com/teilorbarcelos/backend-go/internal/core/repository"
	"gorm.io/gorm"
)

type RoleRepository struct {
	repository.BaseRepository[models.Role]
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		BaseRepository: *repository.NewBaseRepository[models.Role](db),
	}
}

func (r *RoleRepository) FindByID(id string) (*models.Role, error) {
	var role models.Role
	err := r.DB.Preload("RoleFeature").Where("id = ? AND is_deleted = ?", id, false).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) CreateWithPermissions(role *models.Role, permissions []models.RoleFeature) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		for i := range permissions {
			permissions[i].IDRole = role.ID
		}
		if len(permissions) > 0 {
			if err := tx.Create(&permissions).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RoleRepository) UpdateWithPermissions(id string, role *models.Role, permissions []models.RoleFeature) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Role{}).Where("id = ?", id).Updates(role).Error; err != nil {
			return err
		}

		if permissions != nil {
			// Delete existing permissions and recreate (like in Node backend)
			if err := tx.Where("id_role = ?", id).Delete(&models.RoleFeature{}).Error; err != nil {
				return err
			}
			for i := range permissions {
				permissions[i].IDRole = id
			}
			if len(permissions) > 0 {
				if err := tx.Create(&permissions).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

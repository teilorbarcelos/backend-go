package role

import (
	"context"
	"backend-go/internal/core/models"
	"backend-go/internal/infra/session"
	"backend-go/pkg/database"
)

type RoleService struct {
	Repo           *RoleRepository
	SessionManager *session.SessionManager
}

func NewRoleService(repo *RoleRepository, sessionMgr *session.SessionManager) *RoleService {
	return &RoleService{
		Repo:           repo,
		SessionManager: sessionMgr,
	}
}

func (s *RoleService) ListFeatures(ctx context.Context) ([]models.Feature, error) {
	var features []models.Feature
	err := s.Repo.WithContext(ctx).DB.Where("active = ?", true).Find(&features).Error
	return features, err
}

type CreateRoleDTO struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description" binding:"required"`
	Permissions []models.RoleFeature `json:"permissions"`
}

func (s *RoleService) Create(ctx context.Context, dto CreateRoleDTO) (*models.Role, error) {
	role := &models.Role{
		Name:        dto.Name,
		Description: dto.Description,
		Active:      true,
	}
	
	err := s.Repo.WithContext(ctx).CreateWithPermissions(role, dto.Permissions)
	return role, err
}

func (s *RoleService) Update(ctx context.Context, id string, dto CreateRoleDTO) (*models.Role, error) {
	role := &models.Role{
		Name:        dto.Name,
		Description: dto.Description,
	}
	if err := s.Repo.WithContext(ctx).UpdateWithPermissions(id, role, dto.Permissions); err != nil {
		return nil, err
	}
	s.SessionManager.InvalidateRoleSessions(id)
	return s.Repo.WithContext(ctx).FindByID(id)
}

func (s *RoleService) List(ctx context.Context, params database.FilterParams) ([]models.Role, int64, error) {
	allowed := map[string]bool{
		"name":   true,
		"active": true,
	}
	return s.Repo.WithContext(ctx).SearchPaginated(params, allowed)
}

func (s *RoleService) GetByID(ctx context.Context, id string) (*models.Role, error) {
	return s.Repo.WithContext(ctx).FindByID(id)
}

func (s *RoleService) Delete(ctx context.Context, id string) error {
	if err := s.Repo.WithContext(ctx).Delete(id); err != nil {
		return err
	}
	return s.SessionManager.InvalidateRoleSessions(id)
}

func (s *RoleService) SetStatus(ctx context.Context, id string, active bool) error {
	if err := s.Repo.WithContext(ctx).Update(id, map[string]interface{}{"active": active}); err != nil {
		return err
	}
	return s.SessionManager.InvalidateRoleSessions(id)
}

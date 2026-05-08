package role

import (
	"github.com/teilorbarcelos/backend-go/internal/core/models"
	"github.com/teilorbarcelos/backend-go/internal/infra/session"
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

func (s *RoleService) ListFeatures() ([]models.Feature, error) {
	var features []models.Feature
	err := s.Repo.DB.Where("active = ?", true).Find(&features).Error
	return features, err
}

type CreateRoleDTO struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description" binding:"required"`
	Permissions []models.RoleFeature `json:"permissions"`
}

func (s *RoleService) Create(dto CreateRoleDTO) (*models.Role, error) {
	// Em Go, simplificamos o slugify ou apenas usamos o ID fornecido/gerado.
	// O Node usava slugify(name) como ID.
	role := &models.Role{
		Name:        dto.Name,
		Description: dto.Description,
		Active:      true,
	}
	// O BaseModel BeforeCreate vai gerar o UUID se não setarmos ID.
	// Mas o Node usava slugify. Vou gerar um ID baseado no nome se possível para manter consistência.
	// Se preferir UUID, deixe vazio. O Node usava slugify para roles fixas.
	
	err := s.Repo.CreateWithPermissions(role, dto.Permissions)
	return role, err
}

func (s *RoleService) Update(id string, dto CreateRoleDTO) (*models.Role, error) {
	role := &models.Role{
		Name:        dto.Name,
		Description: dto.Description,
	}
	if err := s.Repo.UpdateWithPermissions(id, role, dto.Permissions); err != nil {
		return nil, err
	}
	s.SessionManager.InvalidateRoleSessions(id)
	return s.Repo.FindByID(id)
}

func (s *RoleService) List(offset, limit int) ([]models.Role, int64, error) {
	return s.Repo.FindAll(nil, offset, limit)
}

func (s *RoleService) GetByID(id string) (*models.Role, error) {
	return s.Repo.FindByID(id)
}

func (s *RoleService) Delete(id string) error {
	if err := s.Repo.Delete(id); err != nil {
		return err
	}
	return s.SessionManager.InvalidateRoleSessions(id)
}

func (s *RoleService) SetStatus(id string, active bool) error {
	if err := s.Repo.Update(id, map[string]interface{}{"active": active}); err != nil {
		return err
	}
	return s.SessionManager.InvalidateRoleSessions(id)
}

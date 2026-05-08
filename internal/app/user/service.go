package user

import (
	"errors"

	"github.com/teilorbarcelos/backend-go/internal/core/models"
	"github.com/teilorbarcelos/backend-go/internal/infra/session"
	"github.com/teilorbarcelos/backend-go/pkg/config"
	"github.com/teilorbarcelos/backend-go/pkg/security"
)

type UserService struct {
	Repo           *UserRepository
	SessionManager *session.SessionManager
}

func NewUserService(repo *UserRepository, sessionMgr *session.SessionManager) *UserService {
	return &UserService{
		Repo:           repo,
		SessionManager: sessionMgr,
	}
}

type CreateUserDTO struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	IDRole   string `json:"id_role" binding:"required"`
	Active   bool   `json:"active"`
}

type UpdateUserDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IDRole   string `json:"id_role"`
	Active   *bool  `json:"active"`
}

func (s *UserService) Create(dto CreateUserDTO) (*models.User, error) {
	hashedPassword, err := security.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:   dto.Name,
		Email:  dto.Email,
		Active: true,
		IDRole: dto.IDRole,
		Auth: &models.Auth{
			Password: &hashedPassword,
			Active:   true,
		},
	}

	err = s.Repo.Create(user)
	return user, err
}

func (s *UserService) Update(id string, dto UpdateUserDTO) (*models.User, error) {
	user, err := s.Repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	
	// Proteção: Não permite desativar o primeiro usuário
	if user.Email == config.AppConfig.FirstUserEmail {
		if dto.Active != nil && !*dto.Active {
			return nil, errors.New("o usuário administrador inicial não pode ser desativado")
		}
		// Não permite alterar o email do primeiro usuário para não perder a referência
		if dto.Email != "" && dto.Email != user.Email {
			return nil, errors.New("o email do usuário administrador inicial não pode ser alterado")
		}
	} else {
		if dto.Email != "" {
			updates["email"] = dto.Email
		}
		if dto.Active != nil {
			updates["active"] = *dto.Active
		}
	}

	if dto.Name != "" {
		updates["name"] = dto.Name
	}
	if dto.IDRole != "" {
		updates["id_role"] = dto.IDRole
	}

	if len(updates) > 0 {
		if err := s.Repo.Update(id, updates); err != nil {
			return nil, err
		}
	}

	if dto.Password != "" {
		hashedPassword, _ := security.HashPassword(dto.Password)
		s.Repo.DB.Model(&models.Auth{}).Where("id = ?", user.IDAuth).Update("password", hashedPassword)
	}

	// Invalida sessões se houver qualquer alteração (exceto talvez só nome, mas por segurança invalidamos)
	s.SessionManager.InvalidateUserSessions(id, user.IDRole)

	return s.Repo.FindByID(id)
}

func (s *UserService) List(offset, limit int) ([]models.User, int64, error) {
	return s.Repo.FindAllWithRole(nil, offset, limit)
}

func (s *UserService) GetByID(id string) (*models.User, error) {
	return s.Repo.FindByID(id)
}

func (s *UserService) Delete(id string) error {
	user, err := s.Repo.FindByID(id)
	if err == nil && user.Email == config.AppConfig.FirstUserEmail {
		return errors.New("o usuário administrador inicial não pode ser excluído")
	}
	if err := s.Repo.Delete(id); err != nil {
		return err
	}
	return s.SessionManager.InvalidateUserSessions(id, user.IDRole)
}

func (s *UserService) SetStatus(id string, active bool) error {
	user, err := s.Repo.FindByID(id)
	if err == nil && user.Email == config.AppConfig.FirstUserEmail && !active {
		return errors.New("o usuário administrador inicial não pode ser desativado")
	}
	if err := s.Repo.Update(id, map[string]interface{}{"active": active}); err != nil {
		return err
	}
	return s.SessionManager.InvalidateUserSessions(id, user.IDRole)
}

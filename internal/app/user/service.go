package user

import (
	"context"
	"errors"

	"backend-go/internal/core/models"
	"backend-go/internal/infra/session"
	"backend-go/pkg/config"
	"backend-go/pkg/database"
	"backend-go/pkg/security"
)

type UserRepositoryI interface {
	Create(user *models.User) error
	Update(id string, updates map[string]interface{}) error
	Delete(id string) error
	FindByID(id string, preloads ...string) (*models.User, error)
	FindByEmail(email string, preloads ...string) (*models.User, error)
	UpdatePassword(authID string, password string) error
	SearchPaginated(params database.FilterParams, filterable map[string]database.FilterConfig, searchable []database.SearchConfig, preloads ...string) ([]models.User, int64, error)
	WithContext(ctx context.Context) UserRepositoryI
}

type UserService struct {
	Repo           UserRepositoryI
	SessionManager *session.SessionManager
}

func NewUserService(repo UserRepositoryI, sessionMgr *session.SessionManager) *UserService {
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

func (s *UserService) Create(ctx context.Context, dto CreateUserDTO) (*models.User, error) {
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

	err = s.Repo.WithContext(ctx).Create(user)
	return user, err
}

func (s *UserService) Update(ctx context.Context, id string, dto UpdateUserDTO) (*models.User, error) {
	repo := s.Repo.WithContext(ctx)
	user, err := repo.FindByID(id)
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
		if err := repo.Update(id, updates); err != nil {
			return nil, err
		}
	}

	if dto.Password != "" {
		hashedPassword, _ := security.HashPassword(dto.Password)
		if user.IDAuth != nil {
			if err := repo.UpdatePassword(*user.IDAuth, hashedPassword); err != nil {
				return nil, err
			}
		}
	}

	// Invalida sessões se houver qualquer alteração
	s.SessionManager.InvalidateUserSessions(id, user.IDRole)

	return repo.FindByID(id, "Auth", "Role")
}

func (s *UserService) List(ctx context.Context, params database.FilterParams) ([]models.User, int64, error) {
	// Definimos os campos permitidos para filtro/busca no User seguindo o padrão Node.js
	filterable := map[string]database.FilterConfig{
		"name":      {Operator: "contains"},
		"email":     {Operator: "equals"},
		"active":    {Type: "boolean"},
		"Role.name": {Relation: "nested"},
	}

	searchable := []database.SearchConfig{
		{Key: "name"},
		{Key: "email"},
		{Key: "Role.name", Relation: "nested"},
	}

	return s.Repo.WithContext(ctx).SearchPaginated(params, filterable, searchable, "Role")
}

func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	return s.Repo.WithContext(ctx).FindByID(id, "Auth", "Role")
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	user, err := s.Repo.WithContext(ctx).FindByID(id)
	if err != nil {
		return err
	}
	if user.Email == config.AppConfig.FirstUserEmail {
		return errors.New("o usuário administrador inicial não pode ser excluído")
	}
	if err := s.Repo.WithContext(ctx).Delete(id); err != nil {
		return err
	}
	return s.SessionManager.InvalidateUserSessions(id, user.IDRole)
}

func (s *UserService) SetStatus(ctx context.Context, id string, active bool) error {
	user, err := s.Repo.WithContext(ctx).FindByID(id)
	if err != nil {
		return err
	}
	if user.Email == config.AppConfig.FirstUserEmail && !active {
		return errors.New("o usuário administrador inicial não pode ser desativado")
	}
	if err := s.Repo.WithContext(ctx).Update(id, map[string]interface{}{"active": active}); err != nil {
		return err
	}
	return s.SessionManager.InvalidateUserSessions(id, user.IDRole)
}

package auth

import (
	"context"
	"log"
	"time"

	"backend-go/internal/core/domainerr"
	"backend-go/internal/core/models"
	"backend-go/internal/infra/session"
	"backend-go/pkg/security"
)

type Service interface {
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	GetMe(ctx context.Context, email string) (*LoginResponse, error)
}

type authService struct {
	repo           Repository
	sessionManager session.SessionStore
	GenerateToken  func(id, email, idRole string, permissions []security.Permission) (string, error)
}

func NewService(repo Repository, sessionMgr session.SessionStore) Service {
	return &authService{
		repo:           repo,
		sessionManager: sessionMgr,
		GenerateToken:  security.GenerateToken,
	}
}

type UserResponse struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Phone     *string     `json:"phone"`
	Document  *string     `json:"document"`
	Avatar    *string     `json:"avatar"`
	Active    bool        `json:"active"`
	IDRole    string      `json:"id_role"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	Role      interface{} `json:"role"`
}

type LoginResponse struct {
	Message      string       `json:"message,omitempty"`
	Valid        bool         `json:"valid"`
	Token        string       `json:"token,omitempty"`
	RefreshToken string       `json:"refreshToken,omitempty"`
	User         UserResponse `json:"user"`
}

func (s *authService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domainerr.ErrUserNotFound
	}

	if !user.Active || user.IsDeleted {
		return nil, domainerr.ErrAccountDisabled
	}

	if user.Auth == nil || user.Auth.Password == nil {
		return nil, domainerr.ErrAuthNotConfigured
	}

	if !security.CheckPasswordHash(password, *user.Auth.Password) {
		return nil, domainerr.ErrInvalidCredentials
	}

	return s.prepareAuthResponse(ctx, user)
}

func (s *authService) GetMe(ctx context.Context, email string) (*LoginResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domainerr.ErrUserNotFound
	}

	if !user.Active || user.IsDeleted {
		return nil, domainerr.ErrAccountDisabled
	}

	res, err := s.prepareAuthResponse(ctx, user)
	if err != nil {
		return nil, err
	}
	res.Valid = true
	return res, nil
}

func (s *authService) prepareAuthResponse(ctx context.Context, user *models.User) (*LoginResponse, error) {
	permissions := s.mapPermissions(user)

	token, err := s.GenerateToken(user.ID, user.Email, user.IDRole, permissions)
	if err != nil {
		return nil, domainerr.ErrInternal
	}

	if err := s.createSession(ctx, user, token, permissions); err != nil {
		log.Printf("[AuthService] Erro ao salvar sessão: %v", err)
	}

	return &LoginResponse{
		Valid:        true,
		Token:        token,
		RefreshToken: token, // TODO: Implement proper Refresh Token logic
		User:         s.mapToUserResponse(user, permissions),
	}, nil
}

func (s *authService) mapPermissions(user *models.User) []security.Permission {
	permissions := make([]security.Permission, 0)
	for _, rf := range user.Role.RoleFeature {
		permissions = append(permissions, security.Permission{
			Feature:  rf.IDFeature,
			Create:   rf.Create,
			View:     rf.View,
			Delete:   rf.Delete,
			Activate: rf.Activate,
		})
	}
	return permissions
}

func (s *authService) createSession(ctx context.Context, user *models.User, token string, permissions []security.Permission) error {
	payload := map[string]interface{}{
		"id":          user.ID,
		"email":       user.Email,
		"roleId":      user.IDRole,
		"permissions": permissions,
	}

	tokenHash := security.SHA256(token)
	expireTime := 24 * time.Hour
	refreshExpireTime := 7 * 24 * time.Hour

	if err := s.sessionManager.CreateSession(ctx, user.ID, user.IDRole, tokenHash, payload, expireTime); err != nil {
		return err
	}
	return s.sessionManager.CreateRefreshToken(ctx, user.ID, user.IDRole, tokenHash, refreshExpireTime)
}

func (s *authService) mapToUserResponse(user *models.User, permissions []security.Permission) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Document:  user.Document,
		Avatar:    user.Avatar,
		Active:    user.Active,
		IDRole:    user.IDRole,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Role: map[string]interface{}{
			"id":          user.Role.ID,
			"name":        user.Role.Name,
			"description": user.Role.Description,
			"active":      user.Role.Active,
			"created_at":  user.Role.CreatedAt.Format(time.RFC3339),
			"updated_at":  user.Role.UpdatedAt.Format(time.RFC3339),
			"permissions": permissions,
		},
	}
}

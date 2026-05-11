package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"backend-go/internal/core/models"
	"backend-go/internal/infra/session"
	"backend-go/pkg/security"
)

type AuthServiceI interface {
	Login(email, password string) (*LoginResponse, error)
	GetMe(email string) (*LoginResponse, error)
}

type AuthService struct {
	Repo           AuthRepositoryI
	SessionManager *session.SessionManager
	GenerateToken  func(id, email, idRole string, permissions []security.Permission) (string, error)
}

func NewAuthService(repo AuthRepositoryI, sessionMgr *session.SessionManager) *AuthService {
	return &AuthService{
		Repo:           repo,
		SessionManager: sessionMgr,
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
	RefreshToken string       `json:"refreshToken,omitempty"` // TODO: Implement Refresh Token logic
	User         UserResponse `json:"user"`
}

func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	if !user.Active || user.IsDeleted {
		return nil, errors.New("conta desativada ou removida")
	}

	if user.Auth == nil || user.Auth.Password == nil {
		return nil, errors.New("autenticação não configurada para este usuário")
	}

	if !security.CheckPasswordHash(password, *user.Auth.Password) {
		return nil, errors.New("senha inválida")
	}

	return s.prepareAuthResponse(user)
}

func (s *AuthService) GetMe(email string) (*LoginResponse, error) {
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	if !user.Active || user.IsDeleted {
		return nil, errors.New("conta desativada ou removida")
	}

	res, err := s.prepareAuthResponse(user)
	if err != nil {
		return nil, err
	}
	res.Valid = true
	return res, nil
}

func (s *AuthService) prepareAuthResponse(user *models.User) (*LoginResponse, error) {
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

	token, err := s.GenerateToken(user.ID, user.Email, user.IDRole, permissions)
	if err != nil {
		return nil, err
	}

	refreshToken := token

	payload := map[string]interface{}{
		"id":          user.ID,
		"email":       user.Email,
		"roleId":      user.IDRole,
		"permissions": permissions,
	}

	tokenHash := security.SHA256(token)
	refreshTokenHash := security.SHA256(refreshToken)

	expireTime := 24 * 60 * 60
	refreshExpireTime := 7 * 24 * 60 * 60

	ctx := context.Background()
	if err := s.SessionManager.CreateSession(ctx, user.ID, user.IDRole, tokenHash, payload, time.Duration(expireTime)*time.Second); err != nil {
		log.Printf("[AuthService] Erro ao salvar sessão no Redis: %v", err)
	}
	if err := s.SessionManager.CreateRefreshToken(ctx, user.ID, user.IDRole, refreshTokenHash, time.Duration(refreshExpireTime)*time.Second); err != nil {
		log.Printf("[AuthService] Erro ao salvar refresh token no Redis: %v", err)
	}

	return &LoginResponse{
		Valid:        true,
		Token:        token,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			Document:  user.Document,
			Avatar:    user.Avatar,
			Active:    user.Active,
			IDRole:    user.IDRole,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Role: map[string]interface{}{
				"id":          user.Role.ID,
				"name":        user.Role.Name,
				"description": user.Role.Description,
				"active":      user.Role.Active,
				"created_at":  user.Role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"updated_at":  user.Role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"permissions": permissions,
			},
		},
	}, nil
}

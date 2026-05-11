package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"backend-go/pkg/database"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(email, password string) (*LoginResponse, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockAuthService) GetMe(email string) (*LoginResponse, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		h := NewAuthHandler(mockSvc)
		r := gin.Default()
		r.POST("/login", h.Login)

		res := &LoginResponse{Valid: true, Token: "token"}
		mockSvc.On("Login", "test@test.com", "pass").Return(res, nil)

		body, _ := json.Marshal(LoginRequest{Email: "test@test.com", Password: "pass"})
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		h := NewAuthHandler(mockSvc)
		r := gin.Default()
		r.POST("/login", h.Login)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		h := NewAuthHandler(mockSvc)
		r := gin.Default()
		r.POST("/login", h.Login)

		mockSvc.On("Login", "test@test.com", "pass").Return(nil, errors.New("unauthorized"))

		body, _ := json.Marshal(LoginRequest{Email: "test@test.com", Password: "pass"})
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthHandler_Me(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		h := NewAuthHandler(mockSvc)
		r := gin.Default()
		r.GET("/me", func(c *gin.Context) {
			c.Set("userEmail", "test@test.com")
			h.Me(c)
		})

		res := &LoginResponse{Valid: true}
		mockSvc.On("GetMe", "test@test.com").Return(res, nil)

		req, _ := http.NewRequest("GET", "/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("No Email in Context", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		h := NewAuthHandler(mockSvc)
		r := gin.Default()
		r.GET("/me", h.Me)

		req, _ := http.NewRequest("GET", "/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		h := NewAuthHandler(mockSvc)
		r := gin.Default()
		r.GET("/me", func(c *gin.Context) {
			c.Set("userEmail", "test@test.com")
			h.Me(c)
		})

		mockSvc.On("GetMe", "test@test.com").Return(nil, errors.New("err"))

		req, _ := http.NewRequest("GET", "/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRegisterRoutes(t *testing.T) {
	r := gin.Default()
	db := database.DB
	RegisterRoutes(r.Group(""), r.Group(""), db)
}

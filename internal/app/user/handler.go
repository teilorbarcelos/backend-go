package user

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/teilorbarcelos/backend-go/internal/core/handler"
	"github.com/teilorbarcelos/backend-go/internal/core/models"
	"github.com/teilorbarcelos/backend-go/pkg/database"
)

type UserServiceI interface {
	Create(ctx context.Context, dto CreateUserDTO) (*models.User, error)
	Update(ctx context.Context, id string, dto UpdateUserDTO) (*models.User, error)
	List(ctx context.Context, params database.FilterParams) ([]models.User, int64, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	Delete(ctx context.Context, id string) error
	SetStatus(ctx context.Context, id string, active bool) error
}

type UserHandler struct {
	Service UserServiceI
}

func NewUserHandler(service UserServiceI) *UserHandler {
	return &UserHandler{Service: service}
}

func (h *UserHandler) Create(c *gin.Context) {
	var dto CreateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.Service.Create(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var dto UpdateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.Service.Update(c.Request.Context(), id, dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuário não encontrado"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *UserHandler) List(c *gin.Context) {
	params := handler.ParseFilterParams(c)

	items, total, err := h.Service.List(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
		"page":  params.Page,
		"limit": params.Limit,
	})
}

func (h *UserHandler) ListAll(c *gin.Context) {
	params := handler.ParseFilterParams(c)
	params.Limit = 0 // List all
	params.Filters["ignoreDefaultFilters"] = true

	items, total, err := h.Service.List(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
	})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "usuário excluído com sucesso"})
}

func (h *UserHandler) SetStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.SetStatus(c.Request.Context(), id, body.Active); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status atualizado com sucesso"})
}

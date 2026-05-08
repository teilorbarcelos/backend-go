package role

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/teilorbarcelos/backend-go/internal/core/handler"
)

type RoleHandler struct {
	Service *RoleService
}

func NewRoleHandler(service *RoleService) *RoleHandler {
	return &RoleHandler{Service: service}
}

func (h *RoleHandler) ListFeatures(c *gin.Context) {
	res, err := h.Service.ListFeatures(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var dto CreateRoleDTO
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

func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var dto CreateRoleDTO
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

func (h *RoleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "papel não encontrado"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *RoleHandler) List(c *gin.Context) {
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

func (h *RoleHandler) ListAll(c *gin.Context) {
	params := handler.ParseFilterParams(c)
	params.Limit = 0
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

func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "papel excluído com sucesso"})
}

func (h *RoleHandler) SetStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.SetStatus(c.Request.Context(), id, body.Active); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status atualizado com sucesso"})
}

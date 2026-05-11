package {{.LowerName}}

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"backend-go/internal/core/handler"
)

type {{.Name}}Handler struct {
	Service *{{.Name}}Service
}

func New{{.Name}}Handler(service *{{.Name}}Service) *{{.Name}}Handler {
	return &{{.Name}}Handler{Service: service}
}

func (h *{{.Name}}Handler) Create(c *gin.Context) {
	var entity {{.Name}}
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.Create(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entity)
}

func (h *{{.Name}}Handler) List(c *gin.Context) {
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

func (h *{{.Name}}Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	entity, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *{{.Name}}Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.Update(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *{{.Name}}Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

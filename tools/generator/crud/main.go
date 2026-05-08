package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type TemplateData struct {
	Name      string
	LowerName string
}

const repositoryTemplate = `package {{.LowerName}}

import (
	"backend-go/internal/core/models"
	"backend-go/internal/core/repository"
	"gorm.io/gorm"
)

type {{.Name}} struct {
	models.BaseModel
	Name string ` + "`gorm:\"type:varchar(255);not null\" json:\"name\"`" + `
}

type {{.Name}}Repository struct {
	*repository.BaseRepository[{{.Name}}]
}

func New{{.Name}}Repository(db *gorm.DB) *{{.Name}}Repository {
	return &{{.Name}}Repository{
		BaseRepository: repository.NewBaseRepository[{{.Name}}](db),
	}
}
`

const serviceTemplate = `package {{.LowerName}}

import (
	"context"
	"backend-go/pkg/database"
)

type {{.Name}}Service struct {
	Repo *{{.Name}}Repository
}

func New{{.Name}}Service(repo *{{.Name}}Repository) *{{.Name}}Service {
	return &{{.Name}}Service{Repo: repo}
}

func (s *{{.Name}}Service) Create(ctx context.Context, entity *{{.Name}}) error {
	return s.Repo.WithContext(ctx).Create(entity)
}

func (s *{{.Name}}Service) List(ctx context.Context, params database.FilterParams) ([]{{.Name}}, int64, error) {
	allowed := map[string]bool{
		"name": true,
	}
	return s.Repo.WithContext(ctx).SearchPaginated(params, allowed)
}

func (s *{{.Name}}Service) GetByID(ctx context.Context, id string) (*{{.Name}}, error) {
	return s.Repo.WithContext(ctx).FindByID(id)
}

func (s *{{.Name}}Service) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.Repo.WithContext(ctx).Update(id, updates)
}

func (s *{{.Name}}Service) Delete(ctx context.Context, id string) error {
	return s.Repo.WithContext(ctx).Delete(id)
}
`

const handlerTemplate = `package {{.LowerName}}

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
`

func writeTemplate(path, tmpl string, data TemplateData) {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		log.Fatalf("Erro ao parsear template: %v", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		log.Fatalf("Erro ao executar template: %v", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		log.Fatalf("Erro ao escrever arquivo %s: %v", path, err)
	}
	fmt.Printf("Criado: %s\n", path)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Uso: go run tools/generator/crud/main.go <ModuleName>")
	}

	name := os.Args[1]
	lowerName := strings.ToLower(name)
	
	dir := filepath.Join("internal", "app", lowerName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Erro ao criar diretório %s: %v", dir, err)
	}

	data := TemplateData{
		Name:      name,
		LowerName: lowerName,
	}

	writeTemplate(filepath.Join(dir, "repository.go"), repositoryTemplate, data)
	writeTemplate(filepath.Join(dir, "service.go"), serviceTemplate, data)
	writeTemplate(filepath.Join(dir, "handler.go"), handlerTemplate, data)

	fmt.Printf("\nMódulo '%s' gerado com sucesso em '%s'.\n", name, dir)
}

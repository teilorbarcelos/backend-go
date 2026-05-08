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

const modelTemplate = `package {{.LowerName}}

import "github.com/teilorbarcelos/backend-go/internal/core/models"

type {{.Name}} struct {
	models.BaseModel
	Name string ` + "`gorm:\"type:varchar(255);not null\" json:\"name\"`" + `
}
`

const repositoryTemplate = `package {{.LowerName}}

import (
	"github.com/teilorbarcelos/backend-go/internal/core/repository"
	"gorm.io/gorm"
)

type {{.Name}}Repository struct {
	*repository.BaseRepository[{{.Name}}]
}

func New{{.Name}}Repository(db *gorm.DB) *{{.Name}}Repository {
	return &{{.Name}}Repository{
		BaseRepository: repository.NewBaseRepository[{{.Name}}](db),
	}
}
`

const usecaseTemplate = `package {{.LowerName}}

type {{.Name}}UseCase struct {
	repo *{{.Name}}Repository
}

func New{{.Name}}UseCase(repo *{{.Name}}Repository) *{{.Name}}UseCase {
	return &{{.Name}}UseCase{repo: repo}
}

func (u *{{.Name}}UseCase) Create(entity *{{.Name}}) error {
	return u.repo.Create(entity)
}

func (u *{{.Name}}UseCase) FindAll(filter map[string]interface{}, offset, limit int) ([]{{.Name}}, int64, error) {
	return u.repo.FindAll(filter, offset, limit)
}

func (u *{{.Name}}UseCase) FindByID(id string) (*{{.Name}}, error) {
	return u.repo.FindByID(id)
}

func (u *{{.Name}}UseCase) Update(id string, updates map[string]interface{}) error {
	return u.repo.Update(id, updates)
}

func (u *{{.Name}}UseCase) Delete(id string) error {
	return u.repo.Delete(id)
}
`

const handlerTemplate = `package {{.LowerName}}

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type {{.Name}}Handler struct {
	usecase *{{.Name}}UseCase
}

func New{{.Name}}Handler(usecase *{{.Name}}UseCase) *{{.Name}}Handler {
	return &{{.Name}}Handler{usecase: usecase}
}

func (h *{{.Name}}Handler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/{{.LowerName}}")
	{
		group.POST("/", h.Create)
		group.GET("/", h.FindAll)
		group.GET("/:id", h.FindByID)
		group.PATCH("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}

func (h *{{.Name}}Handler) Create(c *gin.Context) {
	var entity {{.Name}}
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.Create(&entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entity)
}

func (h *{{.Name}}Handler) FindAll(c *gin.Context) {
	// Exemplo simples de extração de offset/limit
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	entities, total, err := h.usecase.FindAll(map[string]interface{}{}, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "total": total})
}

func (h *{{.Name}}Handler) FindByID(c *gin.Context) {
	id := c.Param("id")
	entity, err := h.usecase.FindByID(id)
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

	if err := h.usecase.Update(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *{{.Name}}Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.usecase.Delete(id); err != nil {
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
		log.Fatal("Uso: go run tools/generator/index.go <ModuleName>")
	}

	name := os.Args[1]
	lowerName := strings.ToLower(name)
	
	dir := filepath.Join("internal", "modules", lowerName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Erro ao criar diretório %s: %v", dir, err)
	}

	data := TemplateData{
		Name:      name,
		LowerName: lowerName,
	}

	writeTemplate(filepath.Join(dir, "model.go"), modelTemplate, data)
	writeTemplate(filepath.Join(dir, "repository.go"), repositoryTemplate, data)
	writeTemplate(filepath.Join(dir, "usecase.go"), usecaseTemplate, data)
	writeTemplate(filepath.Join(dir, "handler.go"), handlerTemplate, data)

	fmt.Printf("\nMódulo '%s' gerado com sucesso em '%s'.\n", name, dir)
}

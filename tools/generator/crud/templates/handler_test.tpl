package {{.LowerName}}

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"backend-go/pkg/database"
)

func setup{{.Name}}Handler() (*{{.Name}}Handler, *gin.Engine) {
	repo := New{{.Name}}Repository(database.DB)
	service := New{{.Name}}Service(repo)
	handler := New{{.Name}}Handler(service)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	return handler, r
}

func Test{{.Name}}Handler_Create(t *testing.T) {
	h, r := setup{{.Name}}Handler()
	r.POST("/{{.LowerName}}", h.Create)

	entity := {{.Name}}{
		Name: "Handler Test",
	}
	body, _ := json.Marshal(entity)
	req, _ := http.NewRequest(http.MethodPost, "/{{.LowerName}}", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func Test{{.Name}}Handler_List(t *testing.T) {
	h, r := setup{{.Name}}Handler()
	r.GET("/{{.LowerName}}", h.List)

	req, _ := http.NewRequest(http.MethodGet, "/{{.LowerName}}?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

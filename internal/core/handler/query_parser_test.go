package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParseFilterParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Default values", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", nil)

		params := ParseFilterParams(c)

		assert.Equal(t, 0, params.Pagination.Page)
		assert.Equal(t, 10, params.Pagination.Limit)
		assert.Equal(t, "", params.Order.OrderBy)
		assert.Equal(t, "", params.Order.OrderDirection)
		assert.Empty(t, params.Filters)
	})

	t.Run("Custom pagination and sorting", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/?page=2&limit=50&orderBy=name&orderDirection=desc", nil)

		params := ParseFilterParams(c)

		assert.Equal(t, 2, params.Pagination.Page)
		assert.Equal(t, 50, params.Pagination.Limit)
		assert.Equal(t, "name", params.Order.OrderBy)
		assert.Equal(t, "desc", params.Order.OrderDirection)
	})

	t.Run("Search and filters", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/?searchWord=test&searchFields=name,email&isActive=true&isDeleted=false&role=admin", nil)

		params := ParseFilterParams(c)

		assert.Equal(t, "test", params.SearchWord)
		assert.Equal(t, "name,email", params.SearchFields)
		
		assert.Len(t, params.Filters, 3)
		assert.Equal(t, true, params.Filters["isActive"])
		assert.Equal(t, false, params.Filters["isDeleted"])
		assert.Equal(t, "admin", params.Filters["role"])
	})

	t.Run("Reserved keys should not be in filters", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/?page=1&limit=10&orderBy=id&custom=value", nil)

		params := ParseFilterParams(c)

		assert.Len(t, params.Filters, 1)
		assert.Equal(t, "value", params.Filters["custom"])
		assert.NotContains(t, params.Filters, "page")
		assert.NotContains(t, params.Filters, "limit")
		assert.NotContains(t, params.Filters, "orderBy")
	})
}

package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/teilorbarcelos/backend-go/pkg/database"
)

// ParseFilterParams extrai parâmetros de filtro de uma requisição Gin.
func ParseFilterParams(c *gin.Context) database.FilterParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	params := database.FilterParams{
		Pagination: database.Pagination{
			Page:  page,
			Limit: limit,
		},
		Order: database.Order{
			OrderBy:        c.Query("orderBy"),
			OrderDirection: c.Query("orderDirection"),
		},
		SearchWord:   c.Query("searchWord"),
		SearchFields: c.Query("searchFields"),
		Filters:      make(map[string]interface{}),
	}

	// Captura todos os outros query params como filtros de igualdade ou range
	for key, values := range c.Request.URL.Query() {
		// Pula parâmetros reservados
		if key == "page" || key == "limit" || key == "orderBy" || key == "orderDirection" || key == "searchWord" || key == "searchFields" {
			continue
		}

		if len(values) > 0 {
			val := values[0]
			
			// Conversão básica de tipos (booleanos)
			if val == "true" {
				params.Filters[key] = true
			} else if val == "false" {
				params.Filters[key] = false
			} else {
				params.Filters[key] = val
			}
		}
	}

	return params
}

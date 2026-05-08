package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Pagination contém os parâmetros de paginação.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// Order contém os parâmetros de ordenação.
type Order struct {
	OrderBy        string `json:"orderBy"`
	OrderDirection string `json:"orderDirection"` // "asc" ou "desc"
}

// FilterParams encapsula todos os parâmetros de busca.
type FilterParams struct {
	Pagination
	Order
	SearchWord   string                 `json:"searchWord"`
	SearchFields string                 `json:"searchFields"`
	Filters      map[string]interface{} `json:"filters"`
}

// ApplyFilters aplica filtros dinâmicos a uma query GORM seguindo o padrão do Node.js.
func ApplyFilters(db *gorm.DB, params FilterParams, allowedFilters map[string]bool) *gorm.DB {
	query := db

	// 1. Filtros de Igualdade e Ranges (_start, _end)
	for key, value := range params.Filters {
		if value == nil || value == "" {
			continue
		}

		fieldKey := key
		operator := "="
		
		if strings.HasSuffix(key, "_start") {
			fieldKey = strings.TrimSuffix(key, "_start")
			operator = ">="
		} else if strings.HasSuffix(key, "_end") {
			fieldKey = strings.TrimSuffix(key, "_end")
			operator = "<="
		}

		// Validação básica se o campo é permitido (se fornecido)
		if allowedFilters != nil && !allowedFilters[fieldKey] {
			continue
		}

		// Adiciona a cláusula WHERE
		query = query.Where(fmt.Sprintf("%s %s ?", fieldKey, operator), value)
	}

	// 2. Pesquisa Global (searchWord + searchFields)
	if params.SearchWord != "" && params.SearchFields != "" {
		fields := strings.Split(params.SearchFields, ",")
		var orConditions []string
		var orValues []interface{}

		for _, field := range fields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			
			// Validação se o campo de busca é permitido
			if allowedFilters != nil && !allowedFilters[field] {
				continue
			}

			orConditions = append(orConditions, fmt.Sprintf("%s ILIKE ?", field))
			orValues = append(orValues, "%"+params.SearchWord+"%")
		}

		if len(orConditions) > 0 {
			query = query.Where(strings.Join(orConditions, " OR "), orValues...)
		}
	}

	// 3. Ordenação
	if params.OrderBy != "" {
		direction := "ASC"
		if strings.ToUpper(params.OrderDirection) == "DESC" {
			direction = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", params.OrderBy, direction))
	} else {
		// Ordenação padrão
		query = query.Order("created_at DESC")
	}

	// 4. Paginação
	if params.Limit > 0 {
		page := params.Page
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * params.Limit
		query = query.Offset(offset).Limit(params.Limit)
	}

	return query
}

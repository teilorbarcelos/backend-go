package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type Order struct {
	OrderBy        string `json:"orderBy"`
	OrderDirection string `json:"orderDirection"`
}

type FilterParams struct {
	Pagination
	Order
	SearchWord   string                 `json:"searchWord"`
	SearchFields string                 `json:"searchFields"`
	Filters      map[string]interface{} `json:"filters"`
}

type FilterConfig struct {
	Type      string // "string", "boolean", "date", "number"
	Operator  string // "equals", "contains", "gte", "lte"
	Relation  string // "nested"
	TargetKey string // Nome da coluna no banco se diferente da Key
}

type SearchConfig struct {
	Key      string // Campo de busca (obrigatório pois é slice)
	Relation string // "nested"
}

func ApplyFilters(db *gorm.DB, params FilterParams, filterable map[string]FilterConfig, searchable []SearchConfig) (*gorm.DB, error) {
	query := db
	joinedRelations := make(map[string]bool)
	likeOperator := "ILIKE"

	mainTable := ""
	if query.Statement.Table != "" {
		mainTable = query.Statement.Table
	} else if query.Statement.Model != nil {
		if err := query.Statement.Parse(query.Statement.Model); err == nil {
			mainTable = query.Statement.Schema.Table
		}
	}

	for key, value := range params.Filters {
		if value == nil || value == "" ||
			key == "ignoreDefaultFilters" ||
			key == "page" || key == "limit" || key == "size" ||
			key == "orderBy" || key == "orderDirection" || key == "sort" ||
			key == "searchWord" || key == "searchFields" {
			continue
		}

		fieldKey := key
		operator := ""

		if strings.HasSuffix(key, "_start") {
			fieldKey = strings.TrimSuffix(key, "_start")
			operator = ">="
		} else if strings.HasSuffix(key, "_end") {
			fieldKey = strings.TrimSuffix(key, "_end")
			operator = "<="
		}

		config, ok := filterable[fieldKey]
		if !ok {
			return nil, fmt.Errorf("filtro '%s' não está disponível", fieldKey)
		}

		targetKey := config.TargetKey
		if targetKey == "" {
			targetKey = fieldKey
		}

		if config.Relation == "nested" && strings.Contains(fieldKey, ".") {
			relation := strings.Split(fieldKey, ".")[0]
			if !joinedRelations[relation] {
				query = query.Joins(relation)
				joinedRelations[relation] = true
			}
		}

		if !strings.Contains(targetKey, ".") && mainTable != "" {
			targetKey = fmt.Sprintf("%s.%s", mainTable, targetKey)
		}
		quotedKey := targetKey
		if query.Statement.Schema != nil {
			quotedKey = query.Statement.Quote(targetKey)
		}

		if operator == "" {
			if config.Operator == "contains" {
				operator = likeOperator
				value = "%" + fmt.Sprint(value) + "%"
			} else if config.Operator != "" {
				operator = strings.ToUpper(config.Operator)
				if operator == "EQUALS" {
					operator = "="
				}
			} else {
				operator = "="
			}
		}

		query = query.Where(fmt.Sprintf("%s %s ?", quotedKey, operator), value)
	}

	if params.SearchWord != "" && params.SearchFields != "" {
		fields := strings.Split(params.SearchFields, ",")
		var orConditions []string
		var orValues []interface{}

		for _, requestedField := range fields {
			requestedField = strings.TrimSpace(requestedField)
			if requestedField == "" {
				continue
			}

			var foundConfig *SearchConfig
			for _, s := range searchable {
				if s.Key == requestedField {
					foundConfig = &s
					break
				}
			}

			if foundConfig == nil {
				return nil, fmt.Errorf("campo de busca '%s' não está disponível", requestedField)
			}

			fieldTarget := foundConfig.Key

			if foundConfig.Relation == "nested" && strings.Contains(fieldTarget, ".") {
				relation := strings.Split(fieldTarget, ".")[0]
				if !joinedRelations[relation] {
					query = query.Joins(relation)
					joinedRelations[relation] = true
				}
			}

			if !strings.Contains(fieldTarget, ".") && mainTable != "" {
				fieldTarget = fmt.Sprintf("%s.%s", mainTable, fieldTarget)
			}

			quotedField := fieldTarget
			if query.Statement.Schema != nil {
				quotedField = query.Statement.Quote(fieldTarget)
			}

			orConditions = append(orConditions, fmt.Sprintf("%s %s ?", quotedField, likeOperator))
			orValues = append(orValues, "%"+params.SearchWord+"%")
		}

		if len(orConditions) > 0 {
			query = query.Where(strings.Join(orConditions, " OR "), orValues...)
		}
	}

	if params.OrderBy != "" {
		orderBy := params.OrderBy
		if _, ok := filterable[orderBy]; ok || orderBy == "created_at" || orderBy == "updated_at" {
			if !strings.Contains(orderBy, ".") && mainTable != "" {
				orderBy = fmt.Sprintf("%s.%s", mainTable, orderBy)
			}

			quotedOrder := orderBy
			if query.Statement.Schema != nil {
				quotedOrder = query.Statement.Quote(orderBy)
			}

			direction := "ASC"
			if strings.ToUpper(params.OrderDirection) == "DESC" {
				direction = "DESC"
			}
			query = query.Order(fmt.Sprintf("%s %s", quotedOrder, direction))
		} else {
			return nil, fmt.Errorf("ordenação por '%s' não está disponível", orderBy)
		}
	} else {
		defaultOrder := "created_at"
		if mainTable != "" {
			defaultOrder = fmt.Sprintf("%s.created_at", mainTable)
		}

		if query.Statement.Schema != nil {
			defaultOrder = query.Statement.Quote(defaultOrder)
		}
		query = query.Order(defaultOrder + " DESC")
	}

	if params.Limit > 0 {
		page := params.Page
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * params.Limit
		query = query.Offset(offset).Limit(params.Limit)
	}

	return query, nil
}

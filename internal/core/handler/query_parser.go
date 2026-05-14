package handler

import (
	"strconv"
	"strings"

	"backend-go/pkg/database"

	"github.com/gin-gonic/gin"
)

func ParseFilterParams(c *gin.Context) database.FilterParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	sizeStr := c.Query("size")
	if sizeStr == "" {
		sizeStr = c.DefaultQuery("limit", "25")
	}
	limit, _ := strconv.Atoi(sizeStr)

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

	for key, values := range c.Request.URL.Query() {
		if key == "page" || key == "limit" || key == "size" || key == "orderBy" || key == "orderDirection" || key == "sort" || key == "searchWord" || key == "searchFields" {
			continue
		}

		if len(values) > 0 {
			val := values[0]

			normalizedKey := key
			for _, prefix := range []string{"createdAt", "updatedAt"} {
				if normalizedKey == prefix || normalizedKey == prefix+"_start" || normalizedKey == prefix+"_end" {
					snake := prefix[:7] + "_" + strings.ToLower(prefix[7:])
					normalizedKey = strings.Replace(normalizedKey, prefix, snake, 1)
					break
				}
			}

			switch val {
			case "true":
				params.Filters[normalizedKey] = true
			case "false":
				params.Filters[normalizedKey] = false
			default:
				params.Filters[normalizedKey] = val
			}
		}
	}

	return params
}

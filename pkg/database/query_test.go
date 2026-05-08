package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestModel struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string
	Email string
	Age   int
}

func TestApplyFilters_Functionality(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{DryRun: true})

	t.Run("Basic Equality Filter", func(t *testing.T) {
		params := FilterParams{
			Filters: map[string]interface{}{
				"name": "John",
			},
		}
		query := ApplyFilters(db.Model(&TestModel{}), params, nil)
		sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sql, "name = \"John\"")
	})

	t.Run("Range Filters _start and _end", func(t *testing.T) {
		params := FilterParams{
			Filters: map[string]interface{}{
				"age_start": 20,
				"age_end":   30,
			},
		}
		query := ApplyFilters(db.Model(&TestModel{}), params, nil)
		sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sql, "age >= 20")
		assert.Contains(t, sql, "age <= 30")
	})

	t.Run("Global Search Word", func(t *testing.T) {
		params := FilterParams{
			SearchWord:   "test",
			SearchFields: "name,email",
		}
		query := ApplyFilters(db.Model(&TestModel{}), params, nil)
		sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sql, "name ILIKE \"%test%\" OR email ILIKE \"%test%\"")
	})

	t.Run("Sorting ASC and DESC", func(t *testing.T) {
		paramsASC := FilterParams{Order: Order{OrderBy: "name", OrderDirection: "asc"}}
		queryASC := ApplyFilters(db.Model(&TestModel{}), paramsASC, nil)
		sqlASC := queryASC.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sqlASC, "ORDER BY name ASC")

		paramsDESC := FilterParams{Order: Order{OrderBy: "name", OrderDirection: "desc"}}
		queryDESC := ApplyFilters(db.Model(&TestModel{}), paramsDESC, nil)
		sqlDESC := queryDESC.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sqlDESC, "ORDER BY name DESC")
	})

	t.Run("Pagination", func(t *testing.T) {
		params := FilterParams{Pagination: Pagination{Page: 2, Limit: 10}}
		query := ApplyFilters(db.Model(&TestModel{}), params, nil)
		sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sql, "LIMIT 10 OFFSET 10")
	})
}

func TestApplyFilters_Coverage(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{DryRun: true})

	t.Run("Allowed Filters Validation", func(t *testing.T) {
		allowed := map[string]bool{"name": true}
		params := FilterParams{
			Filters: map[string]interface{}{
				"name":  "John",
				"email": "john@test.com", // Disallowed
			},
		}
		query := ApplyFilters(db.Model(&TestModel{}), params, allowed)
		sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sql, "name = \"John\"")
		assert.NotContains(t, sql, "email")
	})

	t.Run("Search Field Validation and Normalization", func(t *testing.T) {
		allowed := map[string]bool{"name": true}
		params := FilterParams{
			SearchWord:   "test",
			SearchFields: "name, , email", // email disallowed, empty space ignored
		}
		query := ApplyFilters(db.Model(&TestModel{}), params, allowed)
		sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sql, "name ILIKE \"%test%\"")
		assert.NotContains(t, sql, "email")
	})

	t.Run("Ignored Filter Values", func(t *testing.T) {
		params := FilterParams{
			Filters: map[string]interface{}{
				"name":                 "",
				"email":                nil,
				"ignoreDefaultFilters": true,
				"valid":                "ok",
			},
		}
		query := ApplyFilters(db.Model(&TestModel{}), params, nil)
		sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sql, "valid = \"ok\"")
		assert.NotContains(t, sql, "name")
		assert.NotContains(t, sql, "email")
	})

	t.Run("Page Less Than 1", func(t *testing.T) {
		params := FilterParams{Pagination: Pagination{Page: 0, Limit: 10}}
		query := ApplyFilters(db.Model(&TestModel{}), params, nil)
		sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]TestModel{}) })
		assert.Contains(t, sql, "LIMIT 10")
		assert.NotContains(t, sql, "OFFSET")
	})
}

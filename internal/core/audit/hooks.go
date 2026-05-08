package audit

import (
	"fmt"
	"reflect"
	"strings"

	"backend-go/internal/core/models"
	"gorm.io/gorm"
)

// RegisterAuditHooks adiciona callbacks globais no GORM para capturar todas as alterações.
func RegisterAuditHooks(db *gorm.DB) {
	db.Callback().Create().After("gorm:create").Register("audit:create", auditCreateHook)
	db.Callback().Update().Before("gorm:update").Register("audit:update", auditUpdateHook)
	db.Callback().Delete().Before("gorm:delete").Register("audit:delete", auditDeleteHook)
}

func auditCreateHook(db *gorm.DB) {
	if db.Error != nil || db.Statement.Schema == nil || db.Statement.Schema.Table == "audit_log" {
		return
	}

	recordID := getRecordID(db)
	newVals := models.MarshalValues(db.Statement.Dest)

	userID := getUserIDFromContext(db)

	log := models.AuditLog{
		Action:    "CREATE",
		TableName: db.Statement.Schema.Table,
		RecordID:  recordID,
		OldValues: "{}",
		NewValues: newVals,
		UserID:    userID,
	}

	db.Session(&gorm.Session{NewDB: true}).Create(&log)
}

func auditUpdateHook(db *gorm.DB) {
	if db.Error != nil || db.Statement.Schema == nil || db.Statement.Schema.Table == "audit_log" {
		return
	}

	recordID := getRecordID(db)
	if recordID == "" || recordID == "unknown" {
		return
	}

	// Busca valores antigos usando as chaves primárias do GORM
	var oldValues map[string]interface{}
	query := db.Session(&gorm.Session{NewDB: true}).Table(db.Statement.Schema.Table)

	// Aplica as chaves primárias na query
	destValue := reflect.Indirect(reflect.ValueOf(db.Statement.Dest))
	for _, field := range db.Statement.Schema.PrimaryFields {
		var val interface{}
		if destValue.Kind() == reflect.Struct {
			val, _ = field.ValueOf(db.Statement.Context, reflect.ValueOf(db.Statement.Dest))
		} else if destValue.Kind() == reflect.Map {
			// Tenta pegar do mapa
			mapVal := destValue.MapIndex(reflect.ValueOf(field.Name))
			if !mapVal.IsValid() {
				mapVal = destValue.MapIndex(reflect.ValueOf(field.DBName))
			}
			if mapVal.IsValid() {
				val = mapVal.Interface()
			}
		}

		if val != nil {
			query = query.Where(field.DBName+" = ?", val)
		}
	}

	if err := query.Take(&oldValues).Error; err != nil {
		// Se não conseguir pegar o antigo, apenas ignora
		oldValues = make(map[string]interface{})
	}

	newVals := models.MarshalValues(db.Statement.Dest)
	oldValsStr := models.MarshalValues(oldValues)

	userID := getUserIDFromContext(db)

	log := models.AuditLog{
		Action:    "UPDATE",
		TableName: db.Statement.Schema.Table,
		RecordID:  recordID,
		OldValues: oldValsStr,
		NewValues: newVals,
		UserID:    userID,
	}

	// Salva em uma nova sessão para não interferir na transação atual caso ocorra erro
	db.Session(&gorm.Session{NewDB: true}).Create(&log)
}

func auditDeleteHook(db *gorm.DB) {
	if db.Error != nil || db.Statement.Schema == nil || db.Statement.Schema.Table == "audit_logs" {
		return
	}

	recordID := getRecordID(db)
	userID := getUserIDFromContext(db)

	log := models.AuditLog{
		Action:    "DELETE",
		TableName: db.Statement.Schema.Table,
		RecordID:  recordID,
		OldValues: "{}",
		NewValues: "{}",
		UserID:    userID,
	}

	db.Session(&gorm.Session{NewDB: true}).Create(&log)
}

func getRecordID(db *gorm.DB) string {
	if db.Statement.Schema == nil {
		return "unknown"
	}

	destValue := reflect.Indirect(reflect.ValueOf(db.Statement.Dest))
	var ids []string

	for _, field := range db.Statement.Schema.PrimaryFields {
		var val interface{}
		var zero bool

		if destValue.Kind() == reflect.Struct {
			val, zero = field.ValueOf(db.Statement.Context, reflect.ValueOf(db.Statement.Dest))
		} else if destValue.Kind() == reflect.Map {
			mapVal := destValue.MapIndex(reflect.ValueOf(field.Name))
			if !mapVal.IsValid() {
				mapVal = destValue.MapIndex(reflect.ValueOf(field.DBName))
			}
			if mapVal.IsValid() {
				val = mapVal.Interface()
				zero = false
			} else {
				zero = true
			}
		}

		if !zero && val != nil {
			ids = append(ids, fmt.Sprintf("%v", val))
		}
	}

	if len(ids) > 0 {
		return strings.Join(ids, ":")
	}

	return "unknown"
}

func getUserIDFromContext(db *gorm.DB) *string {
	// Pega o UserID do contexto do gin repassado para o GORM (via db.WithContext)
	if ctxVal := db.Statement.Context.Value("userID"); ctxVal != nil {
		if id, ok := ctxVal.(string); ok && id != "" {
			return &id
		}
	}
	return nil
}

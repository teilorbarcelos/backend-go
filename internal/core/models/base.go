package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel é equivalente ao BaseEntity no Node.js.
// Em vez de herança, vamos "embutir" (embed) esta struct nas nossas models.
type BaseModel struct {
	ID        string         `gorm:"type:varchar(40);primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	IsDeleted bool           `gorm:"default:false" json:"is_deleted"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate é um Hook do GORM para garantir que um UUID seja gerado.
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return
}

package entity

import (
	"time"

	"gorm.io/gorm"
)

// Base replaces gorm.Model with proper lowercase JSON tags
type Base struct {
	ID        uint           `gorm:"primarykey"  json:"id"`
	CreatedAt time.Time      `                   json:"createdAt"`
	UpdatedAt time.Time      `                   json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"       json:"-"`
}

package entity

type Role struct {
	Base
	SystemCode  string       `gorm:"not null;index"            json:"systemCode"`
	Name        string       `gorm:"not null"                  json:"name"`
	Description string       `                                  json:"description"`
	IsDefault   bool         `gorm:"default:false"             json:"isDefault"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

package entity

type Menu struct {
	Base
	SystemCode string `gorm:"not null;index"      json:"systemCode"`
	Name       string `gorm:"not null"            json:"name"`
	Code       string `gorm:"not null"            json:"code"`
	Icon       string `                           json:"icon"`
	Path       string `                           json:"path"`
	ParentID   *uint  `                           json:"parentId"`
	SortOrder  int    `gorm:"default:0"           json:"sortOrder"`
	IsActive   bool   `gorm:"default:true"        json:"isActive"`
	Children   []Menu `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

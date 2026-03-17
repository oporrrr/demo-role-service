package entity

type Permission struct {
	Base
	SystemCode  string `gorm:"not null;index" json:"systemCode"`
	Resource    string `gorm:"not null"       json:"resource"`
	Action      string `gorm:"not null"       json:"action"`
	Description string `                      json:"description"`
}

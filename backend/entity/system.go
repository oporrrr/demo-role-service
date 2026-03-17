package entity

type System struct {
	Base
	Code             string `gorm:"uniqueIndex;not null"  json:"code"`
	Name             string `gorm:"not null"              json:"name"`
	Description      string `                             json:"description"`
	APIKey           string `gorm:"uniqueIndex;not null"  json:"-"` // never expose hashed key
	AuthClientID     string `                             json:"-"` // Auth Center client credentials
	AuthClientSecret string `                             json:"-"`
}

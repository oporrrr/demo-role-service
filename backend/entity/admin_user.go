package entity

type AdminUser struct {
	Base
	Username     string `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string `gorm:"not null"             json:"-"`
	DisplayName  string `                            json:"displayName"`
}

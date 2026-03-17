package entity

type UserRole struct {
	Base
	AccountID  string `gorm:"not null;uniqueIndex:idx_user_role_account_system" json:"accountId"`
	RoleID     uint   `gorm:"not null"                                          json:"roleId"`
	Role       Role   `                                                         json:"role"`
	SystemCode string `gorm:"not null;uniqueIndex:idx_user_role_account_system" json:"systemCode"`
}

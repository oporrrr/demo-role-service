package entity

import "time"

// User mirrors the full profile stored in Auth Center.
// AccountID is the primary key — set by Auth Center, not auto-increment.
type User struct {
	AccountID         string    `gorm:"primaryKey"        json:"accountId"`
	KeycloakUserID    string    `gorm:"index"             json:"keycloakUserId"`
	FirstName         string    `                         json:"firstName"`
	LastName          string    `                         json:"lastName"`
	PrefixName        string    `                         json:"prefixName"`
	Gender            string    `                         json:"gender"`
	DateOfBirth       string    `                         json:"dateOfBirth"`
	Email             string    `gorm:"index"             json:"email"`
	Username          string    `                         json:"username"`
	AccountStatus     string    `                         json:"accountStatus"`
	AkidID            int64     `                         json:"akidId"`
	CisNumber         *string   `                         json:"cisNumber"`
	ProfilePictureURL *string   `                         json:"profilePictureUrl"`
	CountryCode       string    `                         json:"countryCode"`
	PhoneNumber       string    `                         json:"phoneNumber"`
	Role              string    `                         json:"role"`
	CreatedAt         time.Time `                         json:"createdAt"`
	UpdatedAt         time.Time `                         json:"updatedAt"`
}

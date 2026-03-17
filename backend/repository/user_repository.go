package repository

import (
	"demo-role-service/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Upsert inserts or updates the user by account_id.
func (r *UserRepository) Upsert(u *entity.User) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"keycloak_user_id",
			"first_name", "last_name", "prefix_name",
			"gender", "date_of_birth",
			"email", "username", "account_status",
			"akid_id", "cis_number", "profile_picture_url",
			"country_code", "phone_number",
			"updated_at",
		}),
	}).Create(u).Error
}

func (r *UserRepository) FindByAccountID(accountID string) (*entity.User, error) {
	var u entity.User
	return &u, r.db.Where("account_id = ?", accountID).First(&u).Error
}

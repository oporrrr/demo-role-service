package repository

import (
	"crypto/sha256"
	"fmt"

	"demo-role-service/entity"
	"gorm.io/gorm"
)

type SystemRepository struct{ db *gorm.DB }

func NewSystemRepository(db *gorm.DB) *SystemRepository { return &SystemRepository{db: db} }

func HashAPIKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}

func (r *SystemRepository) Create(s *entity.System) error {
	return r.db.Create(s).Error
}

func (r *SystemRepository) FindByCode(code string) (*entity.System, error) {
	var s entity.System
	return &s, r.db.Where("code = ?", code).First(&s).Error
}

func (r *SystemRepository) FindByAPIKey(hashed string) (*entity.System, error) {
	var s entity.System
	return &s, r.db.Where("api_key = ?", hashed).First(&s).Error
}

func (r *SystemRepository) List() ([]entity.System, error) {
	var list []entity.System
	return list, r.db.Find(&list).Error
}

func (r *SystemRepository) UpdateAPIKey(code, hashedKey string) error {
	return r.db.Model(&entity.System{}).Where("code = ?", code).Update("api_key", hashedKey).Error
}

func (r *SystemRepository) UpdateCredentials(code, clientID, clientSecret string) error {
	return r.db.Model(&entity.System{}).Where("code = ?", code).
		Updates(map[string]interface{}{"auth_client_id": clientID, "auth_client_secret": clientSecret}).Error
}

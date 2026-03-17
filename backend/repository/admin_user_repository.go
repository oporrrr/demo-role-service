package repository

import (
	"demo-role-service/entity"

	"gorm.io/gorm"
)

type AdminUserRepository struct{ db *gorm.DB }

func NewAdminUserRepository(db *gorm.DB) *AdminUserRepository {
	return &AdminUserRepository{db: db}
}

func (r *AdminUserRepository) Create(u *entity.AdminUser) error {
	return r.db.Create(u).Error
}

func (r *AdminUserRepository) FindByUsername(username string) (*entity.AdminUser, error) {
	var u entity.AdminUser
	return &u, r.db.Where("username = ?", username).First(&u).Error
}

func (r *AdminUserRepository) Count() (int64, error) {
	var count int64
	return count, r.db.Model(&entity.AdminUser{}).Count(&count).Error
}

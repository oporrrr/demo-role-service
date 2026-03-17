package repository

import (
	"demo-role-service/entity"

	"gorm.io/gorm"
)

type MenuRepository struct{ db *gorm.DB }

func NewMenuRepository(db *gorm.DB) *MenuRepository { return &MenuRepository{db: db} }

func (r *MenuRepository) List(systemCode string) ([]entity.Menu, error) {
	var menus []entity.Menu
	return menus, r.db.Where("system_code = ?", systemCode).
		Order("sort_order ASC, id ASC").
		Find(&menus).Error
}

func (r *MenuRepository) Create(m *entity.Menu) error {
	return r.db.Create(m).Error
}

func (r *MenuRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&entity.Menu{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MenuRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Menu{}, id).Error
}

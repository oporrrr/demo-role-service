package repository

import (
	"demo-role-service/entity"
	"gorm.io/gorm"
)

type PermissionRepository struct{ db *gorm.DB }

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) BulkUpsert(perms []entity.Permission) error {
	return r.db.Save(&perms).Error
}

func (r *PermissionRepository) Create(p *entity.Permission) error {
	return r.db.Create(p).Error
}

func (r *PermissionRepository) List(systemCode string) ([]entity.Permission, error) {
	var list []entity.Permission
	q := r.db
	if systemCode != "" {
		q = q.Where("system_code = ?", systemCode)
	}
	return list, q.Find(&list).Error
}

func (r *PermissionRepository) FindByID(id uint) (*entity.Permission, error) {
	var p entity.Permission
	return &p, r.db.First(&p, id).Error
}

func (r *PermissionRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Permission{}, id).Error
}

func (r *PermissionRepository) Exists(systemCode, resource, action string) bool {
	var count int64
	r.db.Model(&entity.Permission{}).
		Where("system_code = ? AND resource = ? AND action = ?", systemCode, resource, action).
		Count(&count)
	return count > 0
}

package repository

import (
	"demo-role-service/entity"
	"gorm.io/gorm"
)

type RoleRepository struct{ db *gorm.DB }

func NewRoleRepository(db *gorm.DB) *RoleRepository { return &RoleRepository{db: db} }

func (r *RoleRepository) Create(role *entity.Role) error {
	return r.db.Create(role).Error
}

func (r *RoleRepository) FindByID(id uint) (*entity.Role, error) {
	var role entity.Role
	return &role, r.db.Preload("Permissions").First(&role, id).Error
}

func (r *RoleRepository) List(systemCode string) ([]entity.Role, error) {
	var list []entity.Role
	q := r.db.Preload("Permissions")
	if systemCode != "" {
		q = q.Where("system_code = ? OR system_code = '*'", systemCode)
	}
	return list, q.Find(&list).Error
}

func (r *RoleRepository) Update(id uint, name, description string) error {
	return r.db.Model(&entity.Role{}).Where("id = ?", id).
		Updates(map[string]any{"name": name, "description": description}).Error
}

func (r *RoleRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Role{}, id).Error
}

// FindDefault returns the default role for a system, or ErrRecordNotFound if none is set.
func (r *RoleRepository) FindDefault(systemCode string) (*entity.Role, error) {
	var list []entity.Role
	if err := r.db.Preload("Permissions").
		Where("system_code = ? AND is_default = true", systemCode).
		Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &list[0], nil
}

// SetDefault marks one role as default and unsets all others in the same system.
func (r *RoleRepository) SetDefault(roleID uint, systemCode string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Role{}).
			Where("system_code = ?", systemCode).
			Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Role{}).
			Where("id = ?", roleID).
			Update("is_default", true).Error
	})
}

func (r *RoleRepository) SetPermissions(roleID uint, permissionIDs []uint) error {
	role := entity.Role{}
	role.ID = roleID
	var perms []entity.Permission
	if len(permissionIDs) > 0 {
		r.db.Find(&perms, permissionIDs)
	}
	return r.db.Model(&role).Association("Permissions").Replace(perms)
}

func (r *RoleRepository) AddPermissions(roleID uint, permissionIDs []uint) error {
	role := entity.Role{}
	role.ID = roleID
	var perms []entity.Permission
	r.db.Find(&perms, permissionIDs)
	return r.db.Model(&role).Association("Permissions").Append(perms)
}

func (r *RoleRepository) RemovePermission(roleID, permissionID uint) error {
	role := entity.Role{}
	role.ID = roleID
	perm := entity.Permission{}
	perm.ID = permissionID
	return r.db.Model(&role).Association("Permissions").Delete(perm)
}

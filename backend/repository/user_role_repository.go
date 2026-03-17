package repository

import (
	"demo-role-service/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRoleRepository struct{ db *gorm.DB }

func NewUserRoleRepository(db *gorm.DB) *UserRoleRepository { return &UserRoleRepository{db: db} }

func (r *UserRoleRepository) Set(accountID, systemCode string, roleID uint) error {
	ur := entity.UserRole{AccountID: accountID, SystemCode: systemCode, RoleID: roleID}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}, {Name: "system_code"}},
		DoUpdates: clause.AssignmentColumns([]string{"role_id", "deleted_at"}),
	}).Create(&ur).Error
}

func (r *UserRoleRepository) GetRoles(accountID string) ([]entity.UserRole, error) {
	var list []entity.UserRole
	return list, r.db.Preload("Role.Permissions").Where("account_id = ?", accountID).Find(&list).Error
}

func (r *UserRoleRepository) GetRoleForSystem(accountID, systemCode string) (*entity.UserRole, error) {
	var list []entity.UserRole
	if err := r.db.Preload("Role.Permissions").
		Where("account_id = ? AND system_code = ?", accountID, systemCode).
		Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &list[0], nil
}

func (r *UserRoleRepository) ListBySystem(systemCode string) ([]entity.UserRole, error) {
	var list []entity.UserRole
	return list, r.db.Unscoped().Preload("Role").Where("system_code = ?", systemCode).Find(&list).Error
}

// IsRemoved returns true when the user had a role that was explicitly revoked (soft-deleted).
func (r *UserRoleRepository) IsRemoved(accountID, systemCode string) bool {
	var count int64
	r.db.Unscoped().Model(&entity.UserRole{}).
		Where("account_id = ? AND system_code = ? AND deleted_at IS NOT NULL", accountID, systemCode).
		Count(&count)
	return count > 0
}

func (r *UserRoleRepository) Remove(accountID, systemCode string) error {
	return r.db.Where("account_id = ? AND system_code = ?", accountID, systemCode).
		Delete(&entity.UserRole{}).Error
}

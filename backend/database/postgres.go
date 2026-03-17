package database

import (
	"log"

	"demo-role-service/config"
	"demo-role-service/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	if config.Cfg.DatabaseURL == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(config.Cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&entity.System{},
		&entity.Role{},
		&entity.Permission{},
		&entity.UserRole{},
		&entity.AdminUser{},
		&entity.Menu{},
		&entity.User{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	// add unique constraint for permission (system_code, resource, action)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_unique ON permissions(system_code, resource, action) WHERE deleted_at IS NULL`)
	// add unique constraint for role (system_code, name)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_unique ON roles(system_code, name) WHERE deleted_at IS NULL`)
	// drop legacy permission column from menus (now derived from menu.code)
	db.Exec(`ALTER TABLE menus DROP COLUMN IF EXISTS permission`)

	DB = db
	log.Println("role-service: database connected and migrated")
}

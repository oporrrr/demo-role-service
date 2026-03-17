package service

import (
	"errors"
	"time"

	"demo-role-service/config"
	"demo-role-service/entity"
	"demo-role-service/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct {
	repo *repository.AdminUserRepository
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{repo: repository.NewAdminUserRepository(db)}
}

// Bootstrap creates the initial "admin" user if no admin exists yet.
// Called on startup when ADMIN_INITIAL_PASSWORD is set.
func (s *AdminService) Bootstrap(password string) error {
	count, _ := s.repo.Count()
	if count > 0 {
		return nil // already have admin users
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.Create(&entity.AdminUser{
		Username:     "admin",
		PasswordHash: string(hash),
		DisplayName:  "Administrator",
	})
}

// Login verifies credentials and returns a signed JWT.
func (s *AdminService) Login(username, password string) (string, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.Username,
		"name": user.DisplayName,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(config.Cfg.JWTSecret))
}

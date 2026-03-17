package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort              string
	DatabaseURL          string
	AdminSecret          string // Bearer token (legacy, still supported)
	RedisURL             string
	JWTSecret            string
	AdminInitialPassword string // seed first admin user on startup if set
}

var Cfg Config

func Load() {
	_ = godotenv.Load()

	Cfg = Config{
		AppPort:              getEnv("APP_PORT", "3001"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		AdminSecret:          getEnv("ADMIN_SECRET", "admin-secret"),
		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:            getEnv("JWT_SECRET", "change-jwt-secret-in-production"),
		AdminInitialPassword: getEnv("ADMIN_INITIAL_PASSWORD", ""),
	}

	log.Printf("config loaded: port=%s", Cfg.AppPort)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

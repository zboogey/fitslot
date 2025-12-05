package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	
	SendGridAPIKey string
	EmailFrom      string
	EmailFromName  string
	RedisAddr      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/fitslot?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "secret-key"),
		
		SendGridAPIKey: getEnv("SENDGRID_API_KEY", ""),
		EmailFrom:      getEnv("EMAIL_FROM", "noreply@fitslot.com"),
		EmailFromName:  getEnv("EMAIL_FROM_NAME", "FitSlot"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
	}
	
	if cfg.JWTSecret == "secret-key" {

	}
	
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

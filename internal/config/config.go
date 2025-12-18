package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string

	EmailFrom     string
	EmailFromName string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	RedisAddr     string
}

func Load() (*Config, error) {
	// We ignore the error because in production/docker we might not have a .env file
	_ = godotenv.Load()

	cfg := &Config{
		Port: getEnv("PORT", "8080"),
		// Use 127.0.0.1 instead of localhost to avoid Mac IPv6 [::1] issues
		DatabaseURL: getEnv("DATABASE_URL", "postgres://testuser:password@127.0.0.1:5434/fitslot_test?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", ""), // Keep empty for validation below

		EmailFrom:     getEnv("EMAIL_FROM", "noreply@fitslot.com"),
		EmailFromName: getEnv("EMAIL_FROM_NAME", "FitSlot"),
		SMTPHost:      getEnv("SMTP_HOST", "127.0.0.1"),
		SMTPPort:      getEnv("SMTP_PORT", "1025"),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPass:      getEnv("SMTP_PASS", ""),
		// Use 127.0.0.1:6380 to match your docker-compose.test.yml
		RedisAddr: getEnv("REDIS_ADDR", "127.0.0.1:6380"),
	}

	// Logic Validation (Criteria 7): Ensure security in production
	if cfg.JWTSecret == "" {
		if os.Getenv("GO_ENV") == "production" {
			return nil, fmt.Errorf("JWT_SECRET must be set in production environment")
		}
		cfg.JWTSecret = "dev-secret-key-change-me"
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	
	// Email configuration (SMTP)
	EmailFrom     string
	EmailFromName string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	RedisAddr     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:admin@localhost:5432/postgres?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "secret-key"),
		
		// Email configuration
		EmailFrom:     getEnv("EMAIL_FROM", "noreply@fitslot.com"),
		EmailFromName: getEnv("EMAIL_FROM_NAME", "FitSlot"),
		SMTPHost:      getEnv("SMTP_HOST", "localhost"),
		SMTPPort:      getEnv("SMTP_PORT", "1025"),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPass:      getEnv("SMTP_PASS", ""),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
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
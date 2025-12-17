package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"fitslot/internal/config"
	"fitslot/internal/db"
	"fitslot/internal/email"
	"fitslot/internal/logger"
	"fitslot/internal/server"
)

func main() {

	logger.Init()
	logger.Info("Starting FitSlot application")
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}
	
	logger.Info("Connecting to database...")
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	logger.Info("Database connected")
	
	if err := db.RunMigrations(database); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Info("Migrations completed")
	
	emailService := email.New(
		cfg.EmailFrom,
		cfg.EmailFromName,
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPass,
		cfg.RedisAddr,
	)
	defer emailService.Close()
	logger.Info("Email service initialized")
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go emailService.Start(ctx)
	
	srv := server.New(database, cfg, emailService)
	
	go func() {
		logger.Infof("Server starting on port %s", cfg.Port)
		if err := srv.Start(cfg.Port); err != nil {
			logger.Fatalf("Server error: %v", err)	
		}
	}()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	
	logger.Info("Shutting down...")
	cancel()
	
	time.Sleep(2 * time.Second)
	
	logger.Info("Server stopped")
}

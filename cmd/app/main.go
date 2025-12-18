package main

import (
	"context"
	_ "fitslot/docs"
	"net/http"
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

// @title FitSlot API
// @version 1.0
// @description API for gym slot booking system.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
		logger.Error("Failed to connect to database", "error", err)
	}
	defer database.Close()
	logger.Info("Database connected")

	if err := db.RunMigrations(database, "migrations"); err != nil {
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

	serverErrChan := make(chan error, 1)
	go func() {
		logger.Infof("Server starting on port %s", cfg.Port)
		if err := srv.Start(cfg.Port); err != nil && err != http.ErrServerClosed {
			serverErrChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Infof("Received signal: %v", sig)
	case err := <-serverErrChan:
		logger.Errorf("Server error: %v", err)
	}

	logger.Info("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Error during server shutdown: %v", err)
	}

	logger.Info("Server stopped")
}

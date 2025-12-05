package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"fitslot/internal/config"
	"fitslot/internal/db"
	"fitslot/internal/email"
	"fitslot/internal/server"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	
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
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go emailService.Start(ctx)
	
	srv := server.New(database, cfg, emailService)
	
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.Start(cfg.Port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down...")
	cancel()
	
	time.Sleep(2 * time.Second)
	
	log.Println("Server stopped")
}
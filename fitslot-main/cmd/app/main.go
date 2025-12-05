package main

import (
	"log"

	"fitslot/internal/config"
	"fitslot/internal/db"
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

	srv := server.New(database, cfg)
	if err := srv.Start(cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}


package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"erp-system/internal/infrastructure/config"
	"erp-system/internal/infrastructure/database/postgres"
	"erp-system/internal/infrastructure/database/seeder"
)

func main() {
	env := flag.String("env", "production", "Seeding mode: production or development")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	adminEmail    := os.Getenv("SEED_ADMIN_EMAIL")
	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")

	if adminEmail == "" {
		adminEmail = "admin@optiluxent.com"
	}
	if adminPassword == "" {
		adminPassword = "Admin@1234"
	}

	db, err := postgres.Connect(cfg)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer postgres.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	switch *env {
	case "development":
		err = seeder.Development(ctx, db, adminEmail, adminPassword)
	default:
		err = seeder.Production(ctx, db, adminEmail, adminPassword)
	}

	if err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	fmt.Printf("✓ Seed complete (mode=%s, admin=%s)\n", *env, adminEmail)
}

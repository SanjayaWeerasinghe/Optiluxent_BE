package main

import (
	"context"
	"fmt"
	"os"

	"erp-system/internal/infrastructure/config"
	"erp-system/internal/infrastructure/database/postgres"
	"erp-system/internal/infrastructure/database/seeder"
)

func main() {
	command := "all"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	db, err := postgres.Connect(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect db: %v\n", err)
		os.Exit(1)
	}
	defer postgres.Close()

	ctx := context.Background()

	adminEmail := os.Getenv("SEED_ADMIN_EMAIL")
	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	if adminEmail == "" {
		adminEmail = "admin@kadahapola.com"
	}
	if adminPassword == "" {
		fmt.Fprintln(os.Stderr, "WARNING: SEED_ADMIN_PASSWORD not set — using insecure default")
		adminPassword = "Admin@12345"
	}

	switch command {
	case "all", "prod", "production":
		if err := seeder.Production(ctx, db, adminEmail, adminPassword); err != nil {
			fmt.Fprintf(os.Stderr, "seed production: %v\n", err)
			os.Exit(1)
		}
	case "dev", "development":
		if err := seeder.Development(ctx, db, adminEmail, adminPassword); err != nil {
			fmt.Fprintf(os.Stderr, "seed development: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s (use: all, prod, dev)\n", command)
		os.Exit(1)
	}

	fmt.Println("Seeding complete.")
}

package main

import (
	stderrors "errors"
	"flag"
	"fmt"
	"os"

	"erp-system/internal/infrastructure/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	command := flag.String("command", "up", "Migration command: up, down, version, force")
	steps := flag.Int("steps", 0, "Number of steps to migrate (up or down). 0 = all.")
	version := flag.Int("version", 0, "Target version for the force command.")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	m, err := migrate.New("file://internal/infrastructure/database/migrations", dsn)
	if err != nil {
		fmt.Printf("Failed to initialise migrate: %v\n", err)
		os.Exit(1)
	}
	defer m.Close()

	switch *command {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
		if err != nil && !stderrors.Is(err, migrate.ErrNoChange) {
			fmt.Printf("Migration up failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
		if err != nil && !stderrors.Is(err, migrate.ErrNoChange) {
			fmt.Printf("Migration down failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations rolled back successfully")

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			fmt.Printf("Failed to get version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Current version: %d  dirty: %v\n", v, dirty)

	case "force":
		if *version == 0 {
			fmt.Println("force requires -version <number>")
			os.Exit(1)
		}
		if err := m.Force(*version); err != nil {
			fmt.Printf("Force failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Forced to version %d\n", *version)

	default:
		fmt.Printf("Unknown command %q. Use: up, down, version, force\n", *command)
		os.Exit(1)
	}
}

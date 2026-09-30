package main

import (
	"flag"
	"fmt"

	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/database"

	"github.com/rs/zerolog/log"
)

func main() {
	var action string
	var name string
	flag.StringVar(&action, "action", "up", "Migration action: up, down, or create")
	flag.StringVar(&name, "name", "", "Migration name (required for action=create)")
	flag.Parse()

	env, err := config.LoadEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env.Database.Postgres.Username,
		env.Database.Postgres.Password,
		env.Database.Postgres.Host,
		env.Database.Postgres.Port,
		env.Database.Postgres.Name,
	)

	switch action {
	case "up":
		log.Info().Msg("Running migrations...")
		if err := database.RunMigrations(databaseURL); err != nil {
			log.Fatal().Err(err).Msg("failed to run migrations")
		}
		log.Info().Msg("Migrations completed successfully")
	case "down":
		log.Info().Msg("Rolling back migrations...")
		if err := database.RollbackMigrations(databaseURL); err != nil {
			log.Fatal().Err(err).Msg("failed to rollback migrations")
		}
		log.Info().Msg("Rollback completed successfully")
	case "create":
		if name == "" {
			log.Fatal().Msg("Migration name is required for 'create' action. Use -name=your_migration_name")
		}
		log.Info().Str("name", name).Msg("Creating new migration files...")
		if err := database.CreateMigration(name); err != nil {
			log.Fatal().Err(err).Msg("failed to create migration")
		}
		log.Info().Msg("Migration files created successfully")
	default:
		log.Fatal().Str("action", action).Msg("invalid action. Use 'up', 'down', or 'create'")
	}
}

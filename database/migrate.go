package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func CreateMigration(name string) error {
	dir := "migrations"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	timestamp := time.Now().Format("20060102150405")
	baseName := fmt.Sprintf("%s_%s", timestamp, name)

	upContent := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n    id SERIAL PRIMARY KEY,\n    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP\n);", name)
	downContent := fmt.Sprintf("DROP TABLE IF EXISTS %s;", name)

	files := map[string]string{
		filepath.Join(dir, baseName+".up.sql"):   upContent,
		filepath.Join(dir, baseName+".down.sql"): downContent,
	}

	for fileName, content := range files {
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create migration file %s: %w", fileName, err)
		}

		_, err = file.WriteString(content)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to write content to %s: %w", fileName, err)
		}
		file.Close()
	}

	fmt.Printf("~ Created migration files for table: %s\n", name)
	return nil
}

func RunMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func RollbackMigrations(databaseURL string) error {
	fmt.Println(databaseURL)
	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return nil
}

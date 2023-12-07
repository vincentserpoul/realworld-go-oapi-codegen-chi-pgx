package database

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // required for migration
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*
var MigrationFiles embed.FS

func Migration(dbURL string) error {
	d, err := iofs.New(MigrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("could not create fs: %w", err)
	}

	m, errM := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if errM != nil {
		return fmt.Errorf("could not migrate conn string: %w", errM)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not migrate up: %w", err)
	}

	return nil
}

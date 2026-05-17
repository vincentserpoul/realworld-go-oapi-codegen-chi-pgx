package database

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // required for migration
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Direction int

const (
	Up Direction = iota
	Down
)

//go:embed migrations/*
var MigrationFiles embed.FS

func Migration(dbURL string, dir Direction) (uint, uint, error) {
	migInst, shutdownFn, err := newMigInstance(dbURL)
	if err != nil {
		return 0, 0, fmt.Errorf("could not migrate, new instance: %w", err)
	}

	defer shutdownFn()

	before, dirty, errV := migInst.Version()
	if errV != nil && !errors.Is(errV, migrate.ErrNilVersion) {
		return 0, 0, fmt.Errorf("could not migrate, before version: %w", errV)
	}

	if dirty {
		return 0, 0, fmt.Errorf("could not migrate, dirty before migration: %w", errV)
	}

	if dir == Down {
		if err := migInst.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return 0, 0, fmt.Errorf("could not migrate down: %w", err)
		}

		return before, 0, nil
	}

	if err := migInst.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return 0, 0, fmt.Errorf("could not migrate up: %w", err)
	}

	after, dirty, errV := migInst.Version()
	if errV != nil {
		return 0, 0, fmt.Errorf("error migration, after version: %w", errV)
	}

	if dirty {
		return 0, 0, fmt.Errorf("error migration, dirty after migration!: %w", errV)
	}

	return before, after, nil
}

func newMigInstance(dbURL string) (*migrate.Migrate, func(), error) {
	drv, err := iofs.New(MigrationFiles, "migrations")
	if err != nil {
		return nil, nil, fmt.Errorf("could not create fs: %w", err)
	}

	migInst, errM := migrate.NewWithSourceInstance("iofs", drv, dbURL)
	if errM != nil && !errors.Is(errM, migrate.ErrNilVersion) {
		return nil, nil, &NewSourceInstanceError{Err: errM}
	}

	const lockTimeout = 30 * time.Second

	// Add a lock timeout of 0 to avoid waiting for the lock
	migInst.LockTimeout = lockTimeout

	shutdownFn := func() {
		if errS, errD := migInst.Close(); errS != nil || errD != nil {
			slog.Error(
				"could not close migration",
				slog.Any("err source", errS),
				slog.Any("err db", errD),
			)
		}

		if errP := drv.Close(); errP != nil {
			slog.Error("could not close source", slog.Any("err", errP))
		}
	}

	return migInst, shutdownFn, nil
}

// db package db impl of repository
package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/induzo/gocom/database/pginit/v2"
	"github.com/induzo/gocom/http/health"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"realworld/internal/repository"
)

type Queryer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// enforce repository interface
var _ repository.Repository = (*Repository)(nil)

type Repository struct {
	pool *pgxpool.Pool
}

var ErrNilLogger = errors.New("logger is nil")

func NewRepository(ctx context.Context, connString string, logger *slog.Logger) (*Repository, error) {
	if logger == nil {
		return nil, ErrNilLogger
	}

	pgi, err := pginit.New(
		connString,
		pginit.WithLogger(logger, "request-id"),
		pginit.WithDecimalType(),
		pginit.WithUUIDType(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PGInit: %w", err)
	}

	pool, err := pgi.ConnPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate connection pool: %w", err)
	}

	return &Repository{
		pool: pool,
	}, nil
}

func (r *Repository) GetShutdownFuncs() map[string]func(ctx context.Context) error {
	return map[string]func(ctx context.Context) error{
		"pgx": func(_ context.Context) error { //nolint:unparam // required by shutdown.Shutdown
			r.pool.Close()

			return nil
		},
	}
}

func (r *Repository) GetHealthChecks() []health.CheckConfig {
	const timeoutPing = 5 * time.Second

	return []health.CheckConfig{
		{
			Name:    "pgx",
			CheckFn: pginit.ConnPoolHealthCheck(r.pool),
			Timeout: timeoutPing,
		},
	}
}

func JSONRowToAddrOfStruct[T any](row pgx.CollectableRow) (*T, error) {
	var dest T

	var jsonBytes []byte
	// scan row into []byte
	if pgxErr := row.Scan(&jsonBytes); pgxErr != nil {
		return nil, fmt.Errorf("could not scan row: %w", pgxErr)
	}

	// unmarshal []byte into struct
	if jsonErr := json.Unmarshal(jsonBytes, &dest); jsonErr != nil {
		return nil, fmt.Errorf("could not unmarshal json: %w", jsonErr)
	}

	return &dest, nil
}

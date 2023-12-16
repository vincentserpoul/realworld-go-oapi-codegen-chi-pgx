// db package db impl of repository
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/induzo/gocom/database/pginit/v2"
	"github.com/induzo/gocom/http/health"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"

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

func NewRepository(ctx context.Context, connString string) (*Repository, error) {
	pgi, err := pginit.New(
		connString,
		pginit.WithTracer(otelpgx.WithTracerProvider(otel.GetTracerProvider())),
		pginit.WithDecimalType(),
		pginit.WithGoogleUUIDType(),
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

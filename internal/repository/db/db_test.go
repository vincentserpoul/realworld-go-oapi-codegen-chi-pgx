package db

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/goleak"

	"realworld/database"
)

var dbURL string //nolint:gochecknoglobals // test code

func TestMain(m *testing.M) {
	ctx := context.Background()

	postgresContainer, errR := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if errR != nil {
		log.Fatalf("could not create container: %v", errR)
	}

	// Clean up the container
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	// get db url
	var errU error

	dbURL, errU = postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if errU != nil {
		log.Fatalf("could not get connection string: %v", errU)
	}

	leak := flag.Bool("leak", false, "use leak detector")
	flag.Parse()

	if *leak {
		goleak.VerifyTestMain(m,
			goleak.IgnoreAnyFunction("github.com/testcontainers/testcontainers-go.(*Reaper).Connect.func1"),
			goleak.IgnoreAnyFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"),
			goleak.IgnoreAnyFunction("database/sql.(*DB).connectionOpener"),
		)

		return
	}

	os.Exit(m.Run())
}

func withRepo(tb testing.TB, dbName string) *Repository {
	tb.Helper()

	ctx := context.Background()

	// create database
	testrep, errC := NewRepository(ctx, dbURL)
	if errC != nil {
		log.Fatalf("could not connect to db: %v", errC)
	}

	if _, err := testrep.pool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName)); err != nil {
		log.Fatalf("could not create test database %s: %v", dbName, err)
	}

	// remove suffix postgres?sslmode=disable and add dbName?sslmode=disable
	newDBURL := strings.ReplaceAll(dbURL, "postgres?sslmode=disable", dbName+"?sslmode=disable")

	if err := database.Migration(newDBURL); err != nil {
		log.Fatalf("could not migrate: %v", err)
	}

	rep, errN := NewRepository(ctx, newDBURL)
	if errN != nil {
		log.Fatalf("could not connect to db: %v", errN)
	}

	return rep
}

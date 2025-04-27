package db

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/induzo/gocom/database/pginit/v2"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/goleak"

	"realworld/database"
)

var dbURL string

func TestMain(m *testing.M) {
	flag.Parse()

	if testing.Short() {
		slog.Warn("skipping pg tests in short mode")
		os.Exit(0)
	}

	ctx := context.Background()

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "17",
		Env: []string{
			"POSTGRES_HOST_AUTH_METHOD=trust",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	resource.Expire(240) // Tell docker to hard kill the container within 4mn

	dbURL = "postgresql://postgres@" + net.JoinHostPort(resource.GetBoundIP("5432/tcp"), resource.GetPort("5432/tcp"))

	if err := pool.Retry(func() error {
		pgi, err := pginit.New(dbURL)
		if err != nil {
			return err
		}

		cPool, errCP := pgi.ConnPool(ctx)
		if errCP == nil {
			cPool.Close()
		}

		return errCP
	}); err != nil {
		log.Fatalf("Could not connect to pg container: %s", err)
	}

	// test migup down once
	if err := runMigDirection(ctx, sourceDB, database.Up); err != nil {
		log.Fatalf("could not create repo: %v", err)
	}

	if err := runMigDirection(ctx, sourceDB, database.Down); err != nil {
		log.Fatalf("could not create repo: %v", err)
	}

	if err := runMigDirection(ctx, sourceDB, database.Up); err != nil {
		log.Fatalf("could not create repo: %v", err)
	}

	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("io.(*pipe).write"),
		goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"),
		goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).triggerHealthCheck"),
		goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).triggerHealthCheck.func1()"),
		goleak.IgnoreTopFunction("database/sql.(*DB).connectionOpener"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).roundTrip"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
	)

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

const sourceDB = "source_db"

func withRepo(tb testing.TB, dbName string) *Repository {
	tb.Helper()

	ctx := tb.Context()

	var newDBURL string

	if strings.Contains(dbURL, "defaultdb") { // when we use the pg testserver
		// remove suffix postgres?sslmode=disable and add dbName?sslmode=disable
		newDBURL = strings.ReplaceAll(dbURL, "defaultdb?sslmode=disable", dbName+"?sslmode=disable")
	} else {
		newDBURL = fmt.Sprintf("%s/%s", dbURL, dbName+"?sslmode=disable")
	}

	if err := runMigDirection(ctx, dbName, database.Up); err != nil {
		log.Fatalf("could not create repo: %v", err)
	}

	rep, errN := NewRepository(tb.Context(), newDBURL)
	if errN != nil {
		tb.Fatalf("could not connect to db: %v", errN)
	}

	return rep
}

func runMigDirection(ctx context.Context, dbName string, dir database.Direction) error {
	// create database
	testrep, errC := NewRepository(ctx, dbURL)
	if errC != nil {
		return fmt.Errorf("could not connect to db: %w", errC)
	}

	if _, err := testrep.pool.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		// if the database already exists, we can ignore the error
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}

		return fmt.Errorf("could not create test database %s: %w", dbName, err)
	}

	var newDBURL string

	if strings.Contains(dbURL, "defaultdb") { // when we use the pg testserver
		// remove suffix postgres?sslmode=disable and add dbName?sslmode=disable
		newDBURL = strings.ReplaceAll(dbURL, "defaultdb?sslmode=disable", dbName+"?sslmode=disable")
	} else {
		newDBURL = fmt.Sprintf("%s/%s", dbURL, dbName+"?sslmode=disable")
	}

	if _, _, err := database.Migration(newDBURL, dir); err != nil {
		return fmt.Errorf("could not migrate: %w", err)
	}

	return nil
}

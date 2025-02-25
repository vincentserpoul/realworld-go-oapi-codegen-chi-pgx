package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"syscall"

	"github.com/induzo/gocom/shutdown"

	"realworld/internal/cmd/api"
	"realworld/internal/domain"
	"realworld/internal/oapi"
	"realworld/internal/repository/db"
)

//nolint:gochecknoglobals // only allowed global vars - filled at build time - do not change
var (
	BuildTime  = "dev"
	CommitHash = "dev"
)

func main() {
	mainCtx, mainStopCtx := context.WithCancel(context.Background())

	environment := "local"
	if os.Getenv("ENV") != "" {
		environment = os.Getenv("ENV")
	}

	cfg, errC := api.ParseConfig(environment)
	if errC != nil {
		log.Fatalf("failed to parse config: %v", errC)
	}

	logger := api.NewLogger(cfg.Log.Level, cfg.Log.IsPretty)

	logger.Info(
		cfg.Name,
		slog.String("buildTime", BuildTime),
		slog.String("commitHash", CommitHash),
		slog.String("env", environment),
	)

	shutdownHandler := shutdown.New(logger)

	server, err := api.NewServer(cfg, shutdownHandler, logger)
	if err != nil {
		log.Fatalf("failed to create api server: %v", err)
	}

	// start otel
	if err := server.StartOtel(); err != nil {
		log.Fatalf("failed to start otel: %v", err)
	}

	// add svc for open api
	repo, errRep := db.NewRepository(mainCtx, cfg.DatabaseURL)
	if errRep != nil {
		log.Fatalf("failed to initiate repository: %v", errRep)
	}

	svc := domain.NewAPISvc(repo)

	// add the openapi http handler and healthchecks on the server
	if err := oapi.RegisterSvc(server, svc, cfg.Security.JWTSecret); err != nil {
		log.Fatalf("failed to register svc: %v", err)
	}

	if err := server.Serve(); err != nil {
		logger.Error(
			"api serve failed with an error",
			slog.Any("err", err),
		)

		os.Exit(1)
	}

	if err := shutdownHandler.Listen(
		mainCtx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	); err != nil {
		logger.Error(
			"graceful shutdown failed.. forcing exit.",
			slog.Any("err", err),
		)

		os.Exit(1)
	}

	mainStopCtx()
}

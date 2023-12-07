package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"syscall"

	"github.com/induzo/gocom/shutdown"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

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

	konf := koanf.New(".")

	if err := konf.Load(file.Provider("config/api/base.toml"), toml.Parser()); err != nil {
		log.Fatalf("failed to load base config: %v", err)
	}

	env := "local"
	if os.Getenv("ENV") != "" {
		env = os.Getenv("ENV")
	}

	if err := konf.Load(file.Provider(fmt.Sprintf("config/api/%s.toml", env)), toml.Parser()); err != nil {
		log.Fatalf("failed to load env config: %v", err)
	}

	cfg := &api.Config{}
	if err := konf.Unmarshal("", &cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	logger := api.NewLogger(cfg.Log.Level, cfg.Log.IsPretty)

	logger.Info("loaded env config", slog.String("env", env))

	shutdownHandler := shutdown.New(logger)

	server, err := api.NewServer(cfg, shutdownHandler, logger)
	if err != nil {
		log.Fatalf("failed to create api server: %v", err)
	}

	// add svc for open api
	repo, errRep := db.NewRepository(mainCtx, cfg.DatabaseURL, logger)
	if errRep != nil {
		log.Fatalf("failed to initiate repository: %v", errRep)
	}

	svc := domain.NewAPISvc(repo)

	// add the openapi http handler and healthchecks on the server
	if err := oapi.RegisterSvc(server, svc, cfg.Security.JWTSecret); err != nil {
		log.Fatalf("failed to register svc: %v", err)
	}

	logger.Info(
		cfg.Name,
		slog.String("buildTime", BuildTime),
		slog.String("commitHash", CommitHash),
	)

	if err := server.Serve(mainCtx); err != nil {
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

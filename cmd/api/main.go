package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"syscall"

	"github.com/induzo/gocom/shutdown"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
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

	environment := "local"
	if os.Getenv("ENV") != "" {
		environment = os.Getenv("ENV")
	}

	cfg, errC := parseConfig(environment)
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

func parseConfig(environment string) (*api.Config, error) {
	konf := koanf.New(".")

	if err := konf.Load(file.Provider("config/api/base.toml"), toml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	if err := konf.Load(file.Provider(fmt.Sprintf("config/api/%s.toml", environment)), toml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load env config from toml: %w", err)
	}

	if err := konf.Load(env.Provider("", ".", func(s string) string {
		return strings.Replace(strings.ToLower(strings.TrimPrefix(s, "")), "_", ".", -1) //nolint:gocritic // magic number
	}), nil); err != nil {
		return nil, fmt.Errorf("failed to load env config from ENV: %w", err)
	}

	cfg := &api.Config{}
	if err := konf.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

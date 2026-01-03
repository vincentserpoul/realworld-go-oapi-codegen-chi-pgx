package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/induzo/gocom/http/health"
	"github.com/induzo/gocom/shutdown"

	"realworld/internal/cmd"
	"realworld/internal/domain"
	"realworld/internal/httpapi"
	"realworld/internal/repository/db"
)

//nolint:gochecknoglobals // only allowed global vars - filled at build time - do not change
var (
	BuildTime  = "dev"
	CommitHash = "dev"
)

const (
	binName = "api"
)

func main() {
	mainCtx, mainStopCtx := context.WithCancel(context.Background())

	environment := cmd.ParseEnvFromString(os.Getenv("ENV"))

	cfg, errC := cmd.ParseConfig[*Config](binName, environment)
	if errC != nil {
		log.Fatalf("failed to parse config: %v", errC)
	}

	logger := cmd.NewLogger(os.Stderr, cfg.Log.Level, cfg.Log.IsPretty)

	logger.InfoContext(
		mainCtx,
		cfg.Name,
		slog.String("buildTime", BuildTime),
		slog.String("commitHash", CommitHash),
		slog.String("env", environment.String()),
	)

	healthchecks := []health.CheckConfig{}

	shutdownHandler := shutdown.New(logger)

	server, err := newAPIServer(mainCtx, logger, healthchecks, shutdownHandler, cfg)
	if err != nil {
		log.Fatalf("newAPIServer: %v", err)
	}

	if err := server.Serve(mainCtx); err != nil {
		log.Fatalf("newAPIServer.Serve: %v", err)
	}

	if err := shutdownHandler.Listen(mainCtx, os.Interrupt); err != nil {
		logger.ErrorContext(
			mainCtx,
			"graceful shutdown failed... forcing exit.",
			slog.Any("err", err),
		)

		os.Exit(1)
	}

	mainStopCtx()
}

type Config struct {
	cmd.BasicConfig `koanf:",squash"`

	DatabaseURL string `koanf:"database_url"`

	Security struct {
		JWTSecret string `koanf:"jwt_secret"`
	} `koanf:"security"`
}

func (cfg *Config) GetBasicConfig() cmd.BasicConfig {
	return cfg.BasicConfig
}

type apiServer struct {
	server          *cmd.Server[*Config]
	cfg             *Config
	shutdownHandler *shutdown.Shutdown
	logger          *slog.Logger
}

func newAPIServer(
	ctx context.Context,
	logger *slog.Logger,
	healthchecks []health.CheckConfig,
	shutdownHandler *shutdown.Shutdown,
	cfg *Config,
) (*apiServer, error) {
	server, err := cmd.NewServer(cfg, shutdownHandler, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create api server: %w", err)
	}

	// new db repository
	rpstry, err := db.NewRepository(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create repo: %w", err)
	}

	healthchecks = append(healthchecks, rpstry.GetHealthChecks()...)

	for _, shut := range rpstry.GetShutdownFuncs() {
		shutdownHandler.Add("pg repository", shut)
	}

	svc := domain.NewAPISvc(rpstry)

	// add the openapi http handler and healthchecks on the server
	rtr, errCR := httpapi.CreateRouter(
		ctx,
		svc,
		logger,
		cfg.WithDebugProfiler,
		cfg.Security.JWTSecret,
	)
	if errCR != nil {
		return nil, fmt.Errorf("failed to create router: %w", errCR)
	}

	// register api router
	if errR := server.RegisterHTTPSvc(
		"/",
		rtr,
		healthchecks,
		map[string]func(ctx context.Context) error{},
	); errR != nil {
		return nil, fmt.Errorf("failed to register http svc: %w", errR)
	}

	return &apiServer{
		server:          server,
		cfg:             cfg,
		shutdownHandler: shutdownHandler,
		logger:          logger,
	}, nil
}

func (api *apiServer) Serve(ctx context.Context) error {
	if err := cmd.StartOtel(ctx, &api.cfg.BasicConfig, api.shutdownHandler, api.logger); err != nil {
		return fmt.Errorf("cmd.StartOtel(): %w", err)
	}

	if err := api.server.Serve(); err != nil {
		return fmt.Errorf("api.server.Serve(): %w", err)
	}

	return nil
}

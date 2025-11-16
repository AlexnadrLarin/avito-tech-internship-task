package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pull-request-service/internal/delivery/http/handlers"
	"pull-request-service/internal/delivery/http/middleware"
	"pull-request-service/internal/delivery/http/routes"
	"pull-request-service/internal/delivery/http/validation"
	"pull-request-service/internal/repository"
	"pull-request-service/internal/service"
	database "pull-request-service/pkg/db"
)

type App struct {
	config *Config
}

func NewApp(config *Config) *App {
	app := &App{
		config: config,
	}

	return app
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logLevel := slog.LevelInfo
	if a.config.LogLevel != "" {
		switch a.config.LogLevel {
		case "ERROR", "error":
			logLevel = slog.LevelError
		case "INFO", "info":
			logLevel = slog.LevelInfo
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	postgres, err := database.NewPostgresWithConfig(ctx, database.PostgresConfig{
		Host:     a.config.Postgres.Host,
		Port:     a.config.Postgres.Port,
		Username: a.config.Postgres.Username,
		Password: a.config.Postgres.Password,
		DBName:   a.config.Postgres.DBName,
		SSLMode:  a.config.Postgres.SSLMode,
		Timeout:  a.config.Postgres.TimeOut,
	})
	if err != nil {
		logger.Error("postgres init failed", "err", err)
		return fmt.Errorf("failed to initialize postgres: %w", err)
	}
	defer postgres.Close()

	txManager := database.NewTransactionManager(postgres.Pool)

	teamsRepository := repository.NewTeamsRepository(postgres.Pool)
	usersRepository := repository.NewUsersRepository(postgres.Pool)
	pullRequestsRepository := repository.NewPullRequestRepository(postgres.Pool)
	reviewRepository := repository.NewReviewRepository(postgres.Pool)

	teamsService := service.NewTeamService(teamsRepository, usersRepository, txManager)
	usersService := service.NewUsersService(usersRepository, reviewRepository, txManager)
	pullRequestService := service.NewPullRequestService(
		pullRequestsRepository,
		reviewRepository,
		teamsRepository,
		txManager,
	)

	validator := validation.NewValidator()

	pullRequestHandler := handlers.NewPullRequestHandler(pullRequestService, logger, validator)
	teamsHandler := handlers.NewTeamHandler(teamsService, logger, validator)
	usersHandler := handlers.NewUsersHandler(usersService, logger, validator)

	loggingMw := middleware.LoggingMiddleware(logger)
	api := routes.SetupMainRouter(loggingMw)

	routes.SetupPullRequestRoutes(api, pullRequestHandler)
	routes.SetupTeamRoutes(api, teamsHandler)
	routes.SetupUsersRoutes(api, usersHandler)

	serverAddr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
	srv := http.Server{
		Addr:    serverAddr,
		Handler: api,
	}

	logger.Info("server started", "addr", serverAddr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server listen failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(
		ctx,
		time.Duration(a.config.Server.ShutdownTimeout)*time.Second,
	)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "err", err)
		return err
	}
	return nil
}

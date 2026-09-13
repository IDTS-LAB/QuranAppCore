package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fiqrikm18/quran-app/internal/core/config"
	"github.com/fiqrikm18/quran-app/internal/core/di"
	"github.com/fiqrikm18/quran-app/internal/core/infrastructure/database"
	"github.com/fiqrikm18/quran-app/internal/core/infrastructure/storage"
	"github.com/fiqrikm18/quran-app/internal/core/middleware"
	"github.com/fiqrikm18/quran-app/internal/core/route"
	"github.com/fiqrikm18/quran-app/internal/core/security"
	"github.com/fiqrikm18/quran-app/internal/core/swagger"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// @title						Quran App API
// @version					1.0
// @description				Core API for the Quran app.
// @BasePath					/api
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Enter "Bearer &lt;access-token&gt;" (get one from the login endpoint).
func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.GetConfig()

	dbConnection, err := database.NewDatabaseConnection(cfg)
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("failed to initialize database connection")
	}

	store, err := storage.New(cfg.S3, logger)
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("failed to initialize s3 storage (is MinIO running? see docker-compose.dev.yml)")
	}

	r := chi.NewRouter()

	// RequestID must come first so every log line and response carries it.
	r.Use(chimw.RequestID)
	r.Use(middleware.ErrorHandler(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(middleware.StructuredLogger(logger))

	container := di.NewContainer(r, dbConnection, logger, security.NewJwtProvider(cfg), store)
	route.RegisterAPI(container)

	// Swagger UI is never exposed in production.
	if swagger.Register(r, cfg.App.Env, logger) {
		logger.Info().Msg("swagger UI enabled at /swagger/index.html")
	}

	server := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,

		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Context is cancelled when SIGINT or SIGTERM is received.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Start HTTP server.
	go func() {
		logger.Info().
			Str("addr", server.Addr).
			Msg("server starting")

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().
				Err(err).
				Msg("server failed")
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()

	logger.Info().
		Msg("shutdown signal received")

	// Give active requests time to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().
			Err(err).
			Msg("server forced to shutdown")
	} else {
		logger.Info().
			Msg("server stopped gracefully")
	}

}

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	corsmiddleware "gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/pkg/cors"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/config"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/service"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/incoming/rest"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/util"
)

func main() {
	if err := run(); err != nil {
		slog.Error("service terminated with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(ctx, ".env")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	util.SetupLogging(cfg)

	greetingService := service.NewGreetingService()

	app, err := rest.NewRestApp(cfg, greetingService)
	if err != nil {
		return fmt.Errorf("creating rest app: %w", err)
	}

	corsMiddleware := corsmiddleware.New(&cfg.CorsConfig)
	handler := corsMiddleware.Handler(app.NewRouter())

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeoutSeconds) * time.Second,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	errChan := make(chan error, 1)
	go func() {
		slog.Info("service listening",
			"addr", server.Addr,
			"stage", cfg.Stage,
			"contextPath", cfg.BasePath,
		)
		if cfg.EnableSwaggerUi && !cfg.IsProd() {
			slog.Info(fmt.Sprintf("swagger ui: http://localhost:%d/%s/swagger-ui", cfg.Port, cfg.BasePath))
		}

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("running http server: %w", err)
			return
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err

	case <-ctx.Done():
		slog.Info("shutdown signal received, draining connections",
			"timeoutSeconds", cfg.ShutdownTimeoutSeconds,
		)
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(cfg.ShutdownTimeoutSeconds)*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutting down http server: %w", err)
		}

		slog.Info("service stopped")
		return nil
	}
}

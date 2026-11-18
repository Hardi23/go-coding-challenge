package config

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	corsmiddleware "gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/pkg/cors"
)

type Config struct {
	Stage                    string                `env:"STAGE, default=dev"`
	BasePath                 string                `env:"BASE_PATH, default=code-challenge"`
	Port                     int                   `env:"PORT, default=8080"`
	LogLevel                 string                `env:"LOG_LEVEL, default=info"`
	EnableSwaggerUi          bool                  `env:"ENABLE_SWAGGER_UI, default=true"`
	MaxEntries               int                   `env:"MAX_ENTRIES, default=10"`
	ReadHeaderTimeoutSeconds int                   `env:"READ_HEADER_TIMEOUT_SECONDS, default=10"`
	ShutdownTimeoutSeconds   int                   `env:"SHUTDOWN_TIMEOUT_SECONDS, default=15"`
	CorsConfig               corsmiddleware.Config `env:",prefix=CORS_"`
}

func (c *Config) IsProd() bool {
	return c.Stage == "prod"
}

func Load(ctx context.Context, envFiles ...string) (*Config, error) {
	if len(envFiles) > 0 {
		err := godotenv.Load(envFiles...)
		if err != nil {
			slog.InfoContext(ctx, "error loading env files", "err", err)
		}
	}

	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("processing env config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535, got %d", c.Port)
	}
	if c.MaxEntries < 1 {
		return fmt.Errorf("MAX_ENTRIES must be greater than 0, got %d", c.MaxEntries)
	}

	trimmedBasePath := strings.TrimSpace(c.BasePath)
	if len(trimmedBasePath) == 0 {
		return fmt.Errorf("BASE_PATH must not be empty")
	}
	if strings.HasPrefix(c.BasePath, "/") || strings.HasSuffix(c.BasePath, "/") {
		return fmt.Errorf("BASE_PATH must not start or end with '/', got %q", c.BasePath)
	}
	switch c.Stage {
	case "dev", "ts", "prod":
	default:
		return fmt.Errorf("STAGE must be one of dev|ts|prod, got %q", c.Stage)
	}
	return nil
}

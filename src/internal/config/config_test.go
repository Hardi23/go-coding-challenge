package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/config"
)

var configEnvKeys = []string{
	"STAGE",
	"BASE_PATH",
	"PORT",
	"LOG_LEVEL",
	"ENABLE_SWAGGER_UI",
	"MAX_ENTRIES",
	"READ_HEADER_TIMEOUT_SECONDS",
	"SHUTDOWN_TIMEOUT_SECONDS",
	"CORS_MAX_AGE",
	"CORS_ALLOW_METHODS",
	"CORS_ALLOW_HEADERS",
	"CORS_ALLOW_CREDENTIALS",
	"CORS_ALLOWED_ORIGIN_REGEX",
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range configEnvKeys {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unsetting %s: %v", key, err)
		}
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearConfigEnv(t)

	cfg, err := config.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "dev", cfg.Stage)
	assert.Equal(t, "code-challenge", cfg.BasePath)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, 10, cfg.MaxEntries)
	assert.False(t, cfg.IsProd())
}

func TestLoad_OverridesFromEnvironment(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("STAGE", "prod")
	t.Setenv("BASE_PATH", "schadenfall-service")
	t.Setenv("PORT", "9090")
	t.Setenv("MAX_ENTRIES", "25")
	t.Setenv("CORS_ALLOW_CREDENTIALS", "true")

	cfg, err := config.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "prod", cfg.Stage)
	assert.Equal(t, "schadenfall-service", cfg.BasePath)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, 25, cfg.MaxEntries)
	assert.True(t, cfg.IsProd())
	assert.True(t, cfg.CorsConfig.CorsAllowCredentials)
}

func TestLoad_RejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "port out of range", key: "PORT", value: "70000"},
		{name: "port zero", key: "PORT", value: "0"},
		{name: "port not a number", key: "PORT", value: "not-a-port"},
		{name: "max entries zero", key: "MAX_ENTRIES", value: "0"},
		{name: "max entries negative", key: "MAX_ENTRIES", value: "-1"},
		{name: "unknown stage", key: "STAGE", value: "staging"},
		{name: "empty context path", key: "BASE_PATH", value: ""},
		{name: "blank context path", key: "BASE_PATH", value: "   "},
		{name: "context path with leading slash", key: "BASE_PATH", value: "/api"},
		{name: "context path with trailing slash", key: "BASE_PATH", value: "api/"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv(tc.key, tc.value)

			cfg, err := config.Load(context.Background())

			require.Error(t, err)
			assert.Nil(t, cfg)
		})
	}
}

func TestLoad_AcceptsAllValidStages(t *testing.T) {
	tests := []struct {
		stage          string
		expectedIsProd bool
	}{
		{stage: "dev", expectedIsProd: false},
		{stage: "ts", expectedIsProd: false},
		{stage: "prod", expectedIsProd: true},
	}

	for _, tc := range tests {
		t.Run(tc.stage, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv("STAGE", tc.stage)

			cfg, err := config.Load(context.Background())

			require.NoError(t, err)
			assert.Equal(t, tc.stage, cfg.Stage)
			assert.Equal(t, tc.expectedIsProd, cfg.IsProd())
		})
	}
}

func TestLoad_RemainingDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg, err := config.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.True(t, cfg.EnableSwaggerUi)
	assert.Equal(t, 10, cfg.ReadHeaderTimeoutSeconds)
	assert.Equal(t, 15, cfg.ShutdownTimeoutSeconds)
	assert.Equal(t, 3600, cfg.CorsConfig.CorsMaxAge)
	assert.False(t, cfg.CorsConfig.CorsAllowCredentials)
	assert.NotEmpty(t, cfg.CorsConfig.CorsAllowOriginRegex)
}

func TestLoad_ReadsEnvFile(t *testing.T) {
	clearConfigEnv(t)

	envFile := filepath.Join(t.TempDir(), ".env")
	content := "STAGE=ts\nBASE_PATH=from-env-file\nPORT=9191\n"
	require.NoError(t, os.WriteFile(envFile, []byte(content), 0o600))

	cfg, err := config.Load(context.Background(), envFile)

	require.NoError(t, err)
	assert.Equal(t, "ts", cfg.Stage)
	assert.Equal(t, "from-env-file", cfg.BasePath)
	assert.Equal(t, 9191, cfg.Port)
}

func TestLoad_EnvFileDoesNotOverrideEnvironment(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("BASE_PATH", "from-environment")

	envFile := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(envFile, []byte("BASE_PATH=from-env-file\n"), 0o600))

	cfg, err := config.Load(context.Background(), envFile)

	require.NoError(t, err)
	assert.Equal(t, "from-environment", cfg.BasePath)
}

func TestLoad_MissingEnvFileIsNotAnError(t *testing.T) {
	clearConfigEnv(t)

	cfg, err := config.Load(context.Background(), filepath.Join(t.TempDir(), "does-not-exist.env"))

	require.NoError(t, err)
	assert.Equal(t, "code-challenge", cfg.BasePath)
}

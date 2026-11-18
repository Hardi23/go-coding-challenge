package util_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/config"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/util"
)

func restoreDefaultLogger(t *testing.T) {
	t.Helper()

	original := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(original)
	})
}

func TestSetupLogging_LogLevelMapping(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     string
		debugEnabled bool
		infoEnabled  bool
		warnEnabled  bool
		errorEnabled bool
	}{
		{
			name:         "debug enables all levels",
			logLevel:     "DEBUG",
			debugEnabled: true,
			infoEnabled:  true,
			warnEnabled:  true,
			errorEnabled: true,
		},
		{
			name:         "info suppresses debug",
			logLevel:     "INFO",
			debugEnabled: false,
			infoEnabled:  true,
			warnEnabled:  true,
			errorEnabled: true,
		},
		{
			name:         "warn suppresses debug and info",
			logLevel:     "WARN",
			debugEnabled: false,
			infoEnabled:  false,
			warnEnabled:  true,
			errorEnabled: true,
		},
		{
			name:         "error suppresses everything below error",
			logLevel:     "ERROR",
			debugEnabled: false,
			infoEnabled:  false,
			warnEnabled:  false,
			errorEnabled: true,
		},
		{
			name:         "lowercase is accepted",
			logLevel:     "warn",
			debugEnabled: false,
			infoEnabled:  false,
			warnEnabled:  true,
			errorEnabled: true,
		},
		{
			name:         "mixed case is accepted",
			logLevel:     "ErRoR",
			debugEnabled: false,
			infoEnabled:  false,
			warnEnabled:  false,
			errorEnabled: true,
		},
		{
			name:         "empty value falls back to debug",
			logLevel:     "",
			debugEnabled: true,
			infoEnabled:  true,
			warnEnabled:  true,
			errorEnabled: true,
		},
		{
			name:         "unknown value falls back to debug",
			logLevel:     "verbose",
			debugEnabled: true,
			infoEnabled:  true,
			warnEnabled:  true,
			errorEnabled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			restoreDefaultLogger(t)

			util.SetupLogging(&config.Config{LogLevel: tc.logLevel})

			ctx := t.Context()
			logger := slog.Default()

			assert.Equal(t, tc.debugEnabled, logger.Enabled(ctx, slog.LevelDebug), "debug")
			assert.Equal(t, tc.infoEnabled, logger.Enabled(ctx, slog.LevelInfo), "info")
			assert.Equal(t, tc.warnEnabled, logger.Enabled(ctx, slog.LevelWarn), "warn")
			assert.Equal(t, tc.errorEnabled, logger.Enabled(ctx, slog.LevelError), "error")
		})
	}
}

func TestSetupLogging_ReplacesDefaultLogger(t *testing.T) {
	restoreDefaultLogger(t)

	before := slog.Default()

	util.SetupLogging(&config.Config{LogLevel: string(util.LogLevelError)})

	assert.NotSame(t, before, slog.Default(), "default logger must be replaced")
}

func TestLogLevelConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, util.LogLevel("DEBUG"), util.LogLevelDebug)
	assert.Equal(t, util.LogLevel("INFO"), util.LogLevelInfo)
	assert.Equal(t, util.LogLevel("WARN"), util.LogLevelWarn)
	assert.Equal(t, util.LogLevel("ERROR"), util.LogLevelError)
}

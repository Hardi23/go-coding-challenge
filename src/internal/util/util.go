package util

import (
	"log/slog"
	"os"
	"strings"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/config"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelError LogLevel = "ERROR"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
)

func SetupLogging(cfg *config.Config) {
	localLogLevel := LogLevelDebug
	if cfg.LogLevel != "" {
		if strings.EqualFold(cfg.LogLevel, string(LogLevelDebug)) {
			localLogLevel = LogLevelDebug
		} else if strings.EqualFold(cfg.LogLevel, string(LogLevelError)) {
			localLogLevel = LogLevelError
		} else if strings.EqualFold(cfg.LogLevel, string(LogLevelInfo)) {
			localLogLevel = LogLevelInfo
		} else if strings.EqualFold(cfg.LogLevel, string(LogLevelWarn)) {
			localLogLevel = LogLevelWarn
		}
	}
	slogLevel := slog.LevelDebug
	switch localLogLevel {
	case LogLevelInfo:
		slogLevel = slog.LevelInfo
	case LogLevelWarn:
		slogLevel = slog.LevelWarn
	case LogLevelError:
		slogLevel = slog.LevelError
	}
	handlerOptions := &slog.HandlerOptions{Level: slogLevel}
	handler := slog.NewJSONHandler(os.Stdout, handlerOptions)
	slog.SetDefault(slog.New(handler))
}

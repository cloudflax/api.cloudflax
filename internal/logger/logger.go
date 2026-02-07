package logger

import (
	"log/slog"
	"os"
)

// Init configures slog as the default logger.
// Uses JSON in production (stdout) and supports LOG_LEVEL (DEBUG, INFO, WARN, ERROR).
// Call at program startup so all logs use this format.
func Init(logLevel string) {
	if logLevel == "" {
		logLevel = "INFO"
	}
	level := parseLevel(logLevel)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	l := slog.New(handler)
	slog.SetDefault(l)
}

func parseLevel(s string) slog.Level {
	switch s {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

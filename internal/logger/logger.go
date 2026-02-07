package logger

import (
	"log/slog"
	"os"
)

// Init configura slog como logger por defecto.
// Usa JSON en producción (stdout) y permite LOG_LEVEL (DEBUG, INFO, WARN, ERROR).
// Llamar al inicio del programa para que todos los logs usen este formato.
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

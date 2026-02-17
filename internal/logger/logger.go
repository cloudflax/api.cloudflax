package logger

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

// Init configures slog as the default logger.
// Uses JSON in production (stdout) and supports LOG_LEVEL (debug, info, warn, error).
// Call at program startup so all logs use this format. Level is output in lowercase.
func Init(logLevel string) {
	if logLevel == "" {
		logLevel = "info"
	}
	level := parseLevel(logLevel)

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				if l, ok := a.Value.Any().(slog.Level); ok {
					return slog.Attr{Key: slog.LevelKey, Value: slog.StringValue(strings.ToLower(l.String()))}
				}
			}
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.Attr{Key: slog.TimeKey, Value: slog.StringValue(t.UTC().Format(time.RFC3339[:19]))}
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	l := slog.New(handler)
	slog.SetDefault(l)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

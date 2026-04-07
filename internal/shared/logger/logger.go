package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Init configures slog as the default logger.
// Uses JSON in production (stdout) and supports LOG_LEVEL (debug, info, warn, error).
// Timestamps are UTC in the form "2006-01-02 15:04:05 UTC" (easier to scan than RFC3339 without zone).
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
				formatted := t.UTC().Format("2006-01-02 15:04:05") + " UTC"
				return slog.Attr{Key: slog.TimeKey, Value: slog.StringValue(formatted)}
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

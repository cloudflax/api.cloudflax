package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Logger logs each request with slog (structured JSON).
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		ip := c.IP()

		err := c.Next()

		attrs := []any{
			"method", method,
			"path", path,
			"ip", ip,
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(start).Milliseconds(),
		}

		if err != nil {
			attrs = append(attrs, "error", err)
			slog.Error("request", attrs...)
			return err
		}

		slog.Info("request", attrs...)
		return nil
	}
}

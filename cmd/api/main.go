package main

import (
	"log/slog"
	"os"

	"github.com/cloudflax/api.cloudflax/internal/app"
	"github.com/cloudflax/api.cloudflax/internal/config"
	"github.com/cloudflax/api.cloudflax/internal/db"
	"github.com/cloudflax/api.cloudflax/internal/logger"
)

func main() {
	logger.Init(os.Getenv("LOG_LEVEL"))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuración inválida", "error", err)
		os.Exit(1)
	}

	if err := db.Init(cfg); err != nil {
		slog.Error("base de datos", "error", err)
		os.Exit(1)
	}

	slog.Info("server starting", "port", cfg.Port)
	if err := app.Run(cfg); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

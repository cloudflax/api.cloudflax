package main

import (
	"log/slog"
	"os"

	"github.com/cloudflax/api.cloudflax/internal/account"
	"github.com/cloudflax/api.cloudflax/internal/app"
	"github.com/cloudflax/api.cloudflax/internal/auth"
	"github.com/cloudflax/api.cloudflax/internal/config"
	"github.com/cloudflax/api.cloudflax/internal/invoice"
	"github.com/cloudflax/api.cloudflax/internal/logger"
	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/user"
)

func main() {
	logger.Init(os.Getenv("LOG_LEVEL"))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	if err := database.Init(cfg); err != nil {
		slog.Error("database", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(&user.User{}, &auth.UserAuthProvider{}, &auth.RefreshToken{}, &account.Account{}, &account.AccountMember{}, &invoice.Invoice{}); err != nil {
		slog.Error("migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("server starting", "port", cfg.Port)
	if err := app.Run(cfg); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

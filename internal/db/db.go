package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cloudflax/api.cloudflax/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Ping verifica que la conexión a PostgreSQL funciona.
func Ping(ctx context.Context, cfg *config.Config) error {
	sslmode := cfg.DBSSLMode
	if sslmode == "" {
		sslmode = "require"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		sslmode,
	)

	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("abrir conexión: %w", err)
	}
	defer conn.Close()

	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	return nil
}

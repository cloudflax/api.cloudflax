package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cloudflax/api.cloudflax/internal/bootstrap/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init opens the PostgreSQL connection.
func Init(cfg *config.Config) error {
	slog.Info("database connecting", "host", cfg.DBHost, "port", cfg.DBPort, "dbname", cfg.DBName, "sslmode", cfg.DBSSLMode)

	dsn := buildDSN(cfg)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open connection: %w", err)
	}

	slog.Info("database connected", "host", cfg.DBHost, "dbname", cfg.DBName, "sslmode", cfg.DBSSLMode)
	return nil
}

// RunMigrations runs AutoMigrate for the given models.
func RunMigrations(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return DB.AutoMigrate(models...)
}

// Ping verifies the PostgreSQL connection is alive.
func Ping(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get connection: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

func buildDSN(cfg *config.Config) string {
	sslmode := cfg.DBSSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		quoteDSNValue(cfg.DBPassword),
		cfg.DBName,
		sslmode,
	)

	if cfg.DBSSLRootCert != "" {
		dsn += fmt.Sprintf(" sslrootcert=%s", cfg.DBSSLRootCert)
	}

	dsn += " connect_timeout=10"

	return dsn
}

// quoteDSNValue wraps a value in single quotes for a libpq key=value DSN,
// escaping embedded single quotes and backslashes.
func quoteDSNValue(v string) string {
	escaped := strings.ReplaceAll(v, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return "'" + escaped + "'"
}

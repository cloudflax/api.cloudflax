package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudflax/api.cloudflax/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init opens the PostgreSQL connection.
func Init(cfg *config.Config) error {
	dsn := buildDSN(cfg)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open connection: %w", err)
	}

	slog.Info("database connected", "host", cfg.DBHost, "dbname", cfg.DBName)
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
		sslmode = "require"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		sslmode,
	)
}

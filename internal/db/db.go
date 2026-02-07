package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudflax/api.cloudflax/internal/config"
	"github.com/cloudflax/api.cloudflax/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init abre la conexión a PostgreSQL y ejecuta las migraciones.
func Init(cfg *config.Config) error {
	dsn := buildDSN(cfg)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("abrir conexión: %w", err)
	}

	if err := RunMigrations(); err != nil {
		return fmt.Errorf("migraciones: %w", err)
	}

	slog.Info("database connected", "host", cfg.DBHost, "dbname", cfg.DBName)
	return nil
}

// RunMigrations ejecuta AutoMigrate para crear/actualizar tablas.
func RunMigrations() error {
	return DB.AutoMigrate(
		&models.User{},
	)
}

// Ping verifica que la conexión a PostgreSQL funciona.
func Ping(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("base de datos no inicializada")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("obtener conexión: %w", err)
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

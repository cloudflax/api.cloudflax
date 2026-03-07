// Package main provides a development-only CLI to truncate all application tables.
// It must not be run in production; it refuses to run when APP_ENV=production (or prod).
// DB credentials are resolved from AWS Secrets Manager (same flow as the main API).
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/secrets"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "production" || env == "prod" {
		fmt.Fprintln(os.Stderr, "db-reset is for development only. Refusing to run when APP_ENV=production.")
		os.Exit(1)
	}

	secretName := os.Getenv("AWS_SECRET_NAME")
	if secretName == "" {
		fmt.Fprintln(os.Stderr, "AWS_SECRET_NAME is required (Secrets Manager secret name)")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	creds, err := fetchDBCredentials(ctx, secretName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch db credentials: %v\n", err)
		os.Exit(1)
	}

	dsn := buildDSN(creds)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "database connection: %v\n", err)
		os.Exit(1)
	}

	sql := `TRUNCATE TABLE refresh_tokens, user_auth_providers, account_members, invoices, accounts, users RESTART IDENTITY CASCADE`
	if err := db.Exec(sql).Error; err != nil {
		fmt.Fprintf(os.Stderr, "truncate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Database tables truncated successfully.")
}

func fetchDBCredentials(ctx context.Context, secretName string) (*secrets.DBCredentials, error) {
	provider, err := secrets.NewSecretsManagerProvider(ctx, secrets.SecretsManagerOptions{
		EndpointURL:     os.Getenv("AWS_ENDPOINT_URL"),
		Region:          os.Getenv("AWS_REGION"),
		SecretID:        secretName,
		Profile:         os.Getenv("AWS_PROFILE"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	})
	if err != nil {
		return nil, fmt.Errorf("create secrets provider: %w", err)
	}
	return provider.GetDBCredentials(ctx)
}

func buildDSN(creds *secrets.DBCredentials) string {
	sslmode := os.Getenv("DB_SSL_MODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		creds.Host(),
		creds.Port(),
		creds.Username(),
		quoteDSNValue(creds.Password()),
		creds.DBName(),
		sslmode,
	)

	if cert := os.Getenv("DB_SSL_ROOT_CERT"); cert != "" {
		dsn += fmt.Sprintf(" sslrootcert=%s", cert)
	}

	dsn += " connect_timeout=10"
	return dsn
}

func quoteDSNValue(v string) string {
	escaped := strings.ReplaceAll(v, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return "'" + escaped + "'"
}

// Package main provides a development-only CLI to truncate all application tables.
// It must not be run in production; it refuses to run when APP_ENV=production (or prod).
// It reads DB connection from environment (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSL_MODE).
package main

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "production" || env == "prod" {
		fmt.Fprintln(os.Stderr, "db-reset is for development only. Refusing to run when APP_ENV=production.")
		os.Exit(1)
	}

	dsn := buildDSN()
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

func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "cloudflax")
	sslmode := getEnv("DB_SSL_MODE", "disable")
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

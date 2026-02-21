package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/secrets"
)

// Config holds the application configuration.
type Config struct {
	Port      string
	LogLevel  string // debug, info, warn, error
	JWTSecret string
	AppURL    string

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string // require, verify-ca, verify-full, disable

	AWSRegion          string
	AWSEndpointURL     string
	AWSAccessKeyID     string
	AWSSecretAccessKey string

	SESFromAddress string
}

var (
	dbSecretsProvider secrets.Provider
	dbSecretsOnce     sync.Once
	dbSecretsInitErr  error
)

// Load loads the configuration. DB credentials always come from AWS Secrets Manager
// (e.g. LocalStack). Server settings (PORT, LOG_LEVEL) and DB_SSL_MODE always come
// from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", ""),
		LogLevel:           getEnv("LOG_LEVEL", ""),
		DBSSLMode:          getEnv("DB_SSL_MODE", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		AppURL:             getEnv("APP_URL", ""),
		AWSRegion:          getEnv("AWS_REGION", ""),
		AWSEndpointURL:     awsEndpointURL(),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		SESFromAddress:     getEnv("SES_FROM_ADDRESS", ""),
	}

	secretName := getEnv("AWS_SECRET_NAME", "")
	if secretName == "" {
		return nil, fmt.Errorf("AWS_SECRET_NAME is required (Secrets Manager secret name)")
	}
	dbCfg, err := loadDBConfigFromSecrets(context.Background(), secretName)
	if err != nil {
		return nil, fmt.Errorf("load db config from secrets: %w", err)
	}
	cfg.DBHost = dbCfg.Host()
	cfg.DBPort = dbCfg.Port()
	cfg.DBUser = dbCfg.Username()
	cfg.DBPassword = dbCfg.Password()
	cfg.DBName = dbCfg.DBName()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadDBConfigFromSecrets fetches DB credentials from AWS Secrets Manager (e.g. LocalStack) at startup.
// secretName is the name or ARN of the secret in Secrets Manager (provided by the deployer).
func loadDBConfigFromSecrets(ctx context.Context, secretName string) (*secrets.DBCredentials, error) {
	dbSecretsOnce.Do(func() {
		ttlSeconds := getEnvInt("SECRETS_CACHE_TTL_SECONDS", 300)

		baseProvider, err := secrets.NewSecretsManagerProvider(ctx, secrets.SecretsManagerOptions{
			EndpointURL:     awsEndpointURL(),
			Region:          getEnv("AWS_REGION", ""),
			SecretID:        secretName,
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			CacheTTL:        time.Duration(ttlSeconds) * time.Second,
		})
		if err != nil {
			dbSecretsInitErr = err
			return
		}
		dbSecretsProvider = baseProvider
	})

	if dbSecretsInitErr != nil {
		return nil, dbSecretsInitErr
	}
	if dbSecretsProvider == nil {
		return nil, fmt.Errorf("secrets provider not initialized")
	}

	return dbSecretsProvider.GetDBCredentials(ctx)
}

// Validate verifies that required configuration is present.
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

// awsEndpointURL resolves the AWS endpoint URL. It reads AWS_ENDPOINT_URL
// first. When APP_ENV=localstack and no explicit URL is set, it falls back to
// LOCALSTACK_ENDPOINT (required in that environment).
func awsEndpointURL() string {
	if v := os.Getenv("AWS_ENDPOINT_URL"); v != "" {
		return v
	}

	if getEnv("APP_ENV", "") == "localstack" {
		return getEnv("LOCALSTACK_ENDPOINT", "")
	}

	return ""
}

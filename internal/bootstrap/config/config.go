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

	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string // require, verify-ca, verify-full, disable
	DBSSLRootCert string // path to CA certificate (e.g. /certs/global-bundle.pem)

	AWSRegion          string
	AWSProfile         string // named profile from ~/.aws/config (e.g. "dev")
	AWSEndpointURL     string
	AWSAccessKeyID     string
	AWSSecretAccessKey string

	SESFromAddress  string
	SESEndpointURL  string
}

var (
	dbSecretsProvider secrets.Provider
	dbSecretsOnce     sync.Once
	dbSecretsInitErr  error
)

// Load loads the configuration. DB credentials come from AWS Secrets Manager.
// Server settings (PORT, LOG_LEVEL) and DB_SSL_MODE come from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", ""),
		LogLevel:           getEnv("LOG_LEVEL", ""),
		DBSSLMode:          getEnv("DB_SSL_MODE", ""),
		DBSSLRootCert:      getEnv("DB_SSL_ROOT_CERT", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		AppURL:             getEnv("APP_URL", ""),
		AWSRegion:          getEnv("AWS_REGION", ""),
		AWSProfile:         getEnv("AWS_PROFILE", ""),
		AWSEndpointURL:     awsEndpointURL(),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		SESFromAddress:     getEnv("SES_FROM_ADDRESS", ""),
		SESEndpointURL:     getEnv("SES_ENDPOINT_URL", ""),
	}

	secretName := getEnv("AWS_SECRET_NAME", "")
	if secretName == "" {
		return nil, fmt.Errorf("AWS_SECRET_NAME is required (Secrets Manager secret name)")
	}

	secretsCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	dbCfg, err := loadDBConfigFromSecrets(secretsCtx, secretName)
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

// loadDBConfigFromSecrets fetches DB credentials from AWS Secrets Manager at startup.
// secretName is the name or ARN of the secret in Secrets Manager.
func loadDBConfigFromSecrets(ctx context.Context, secretName string) (*secrets.DBCredentials, error) {
	dbSecretsOnce.Do(func() {
		ttlSeconds := getEnvInt("SECRETS_CACHE_TTL_SECONDS", 300)

		baseProvider, err := secrets.NewSecretsManagerProvider(ctx, secrets.SecretsManagerOptions{
			EndpointURL:     awsEndpointURL(),
			Region:          getEnv("AWS_REGION", ""),
			SecretID:        secretName,
			Profile:         getEnv("AWS_PROFILE", ""),
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
	if (c.DBSSLMode == "verify-full" || c.DBSSLMode == "verify-ca") && c.DBSSLRootCert == "" {
		return fmt.Errorf("DB_SSL_ROOT_CERT is required when DB_SSL_MODE is %s", c.DBSSLMode)
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

// awsEndpointURL resolves a custom AWS endpoint URL from the environment.
// Leave AWS_ENDPOINT_URL empty to use the standard AWS endpoints.
func awsEndpointURL() string {
	if v := os.Getenv("AWS_ENDPOINT_URL"); v != "" {
		return v
	}
	return ""
}

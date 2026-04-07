package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/secrets"
)

// Config holds the application configuration.
type Config struct {
	Port        string
	LogLevel    string // debug, info, warn, error
	AppEnv      string
	JWTSecret   string
	AppURL      string
	FrontendURL string

	// EnableAuthDevEndpoints exposes POST /auth/dev/* helpers when true and AppEnv is not production.
	EnableAuthDevEndpoints bool

	// TrustProxy enables reading client IP from ProxyHeader when the direct peer is a trusted proxy.
	TrustProxy           bool
	ProxyHeader          string
	TrustProxyPrivate    bool
	TrustProxyLoopback   bool

	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string // require, verify-ca, verify-full, disable
	DBSSLRootCert string // path to CA certificate (e.g. /certs/global-bundle.pem)
	// DBSlowQueryThresholdMS is the slow-query log threshold in milliseconds.
	DBSlowQueryThresholdMS int

	AWSRegion          string
	AWSProfile         string // named profile from ~/.aws/config (e.g. "dev")
	AWSEndpointURL     string
	AWSAccessKeyID     string
	AWSSecretAccessKey string

	SESFromAddress string
	SESEndpointURL string

	// Verification email is sent by Lambda (async).
	LambdaSendVerifyEmailName string
	// Forgot-password email is sent by a dedicated Lambda (async), same invocation style as verification.
	LambdaSendForgotPasswordEmailName string
	APIThrottleTableName              string

	// JWTAccessTokenDuration is the signed JWT access token lifetime.
	JWTAccessTokenDuration time.Duration
}

var (
	dbSecretsProvider secrets.Provider
	dbSecretsOnce     sync.Once
	dbSecretsInitErr  error
)

// Load loads the configuration. DB credentials come from AWS Secrets Manager.
// Server settings (PORT, LOG_LEVEL) and DB_SSL_MODE come from environment variables.
func Load() (*Config, error) {
	proxyHeader := strings.TrimSpace(getEnv("PROXY_HEADER", "X-Forwarded-For"))
	if proxyHeader == "" {
		proxyHeader = "X-Forwarded-For"
	}
	appEnv := strings.TrimSpace(getEnv("APP_ENV", ""))
	cfg := &Config{
		Port:                      getEnv("PORT", ""),
		LogLevel:                  getEnv("LOG_LEVEL", ""),
		AppEnv:                    appEnv,
		DBSSLMode:                 getEnv("DB_SSL_MODE", ""),
		DBSSLRootCert:             getEnv("DB_SSL_ROOT_CERT", ""),
		DBSlowQueryThresholdMS:    resolveSlowQueryThresholdMS(),
		JWTSecret:                 getEnv("JWT_SECRET", ""),
		AppURL:                    getEnv("APP_URL", ""),
		FrontendURL:               getEnv("FRONTEND_URL", ""),
		EnableAuthDevEndpoints:    envBool("ENABLE_AUTH_DEV_ENDPOINTS", false),
		TrustProxy:                envBool("TRUST_PROXY", false),
		ProxyHeader:               proxyHeader,
		TrustProxyPrivate:         envBool("TRUST_PROXY_TRUST_PRIVATE", true),
		TrustProxyLoopback:        envBool("TRUST_PROXY_TRUST_LOOPBACK", false),
		AWSRegion:                 getEnv("AWS_REGION", ""),
		AWSProfile:                getEnv("AWS_PROFILE", ""),
		AWSEndpointURL:            awsEndpointURL(),
		AWSAccessKeyID:            getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:        getEnv("AWS_SECRET_ACCESS_KEY", ""),
		SESFromAddress:            getEnv("SES_FROM_ADDRESS", ""),
		SESEndpointURL:            getEnv("SES_ENDPOINT_URL", ""),
		LambdaSendVerifyEmailName:             getEnv("LAMBDA_SEND_VERIFY_EMAIL_NAME", ""),
		LambdaSendForgotPasswordEmailName:     getEnv("LAMBDA_SEND_FORGOT_PASSWORD_EMAIL_NAME", ""),
		APIThrottleTableName:                  getEnv("API_THROTTLE_TABLE_NAME", ""),
		JWTAccessTokenDuration:    jwtAccessTokenDurationFromEnv(),
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
	if c.JWTAccessTokenDuration < time.Minute {
		return fmt.Errorf("JWT_ACCESS_TOKEN_DURATION_MINUTES must be at least 1")
	}
	if c.JWTAccessTokenDuration > 7*24*time.Hour {
		return fmt.Errorf("JWT_ACCESS_TOKEN_DURATION_MINUTES must not exceed 10080 (7 days)")
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

// resolveSlowQueryThresholdMS returns DB_SLOW_QUERY_THRESHOLD_MS when set to a positive value;
// otherwise 500 in development (APP_ENV=development) and 200 for other environments.
func resolveSlowQueryThresholdMS() int {
	if ms := getEnvInt("DB_SLOW_QUERY_THRESHOLD_MS", 0); ms > 0 {
		return ms
	}
	if getEnv("APP_ENV", "") == "development" {
		return 500
	}
	return 200
}

// awsEndpointURL resolves a custom AWS endpoint URL from the environment.
// Leave AWS_ENDPOINT_URL empty to use the standard AWS endpoints.
func awsEndpointURL() string {
	if v := os.Getenv("AWS_ENDPOINT_URL"); v != "" {
		return v
	}
	return ""
}

// jwtAccessTokenDurationFromEnv reads JWT_ACCESS_TOKEN_DURATION_MINUTES (default 15).
func jwtAccessTokenDurationFromEnv() time.Duration {
	mins := getEnvInt("JWT_ACCESS_TOKEN_DURATION_MINUTES", 15)
	return time.Duration(mins) * time.Minute
}

// envBool parses common truthy/falsey environment strings; empty uses defaultVal.
func envBool(key string, defaultVal bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return defaultVal
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultVal
	}
}

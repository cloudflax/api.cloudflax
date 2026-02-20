package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/v2/secretcache"
)

// Provider returns database credentials from an external source (e.g. AWS Secrets Manager).
type Provider interface {
	GetDBCredentials(ctx context.Context) (*DBCredentials, error)
}

// secretCacheAPI is a minimal interface for the Secrets Manager secret cache.
// It is used so the cache can be stubbed in tests.
type secretCacheAPI interface {
	GetSecretStringWithContext(ctx context.Context, secretId string) (string, error)
}

// SecretsManagerProvider loads DBCredentials from AWS Secrets Manager (LocalStack or AWS).
type SecretsManagerProvider struct {
	cache  secretCacheAPI
	secret string
}

// SecretsManagerOptions configures the Secrets Manager provider.
type SecretsManagerOptions struct {
	EndpointURL     string
	Region          string
	SecretID        string
	AccessKeyID     string
	SecretAccessKey string
	CacheTTL        time.Duration
}

// NewSecretsManagerProvider creates a provider that fetches DB credentials from Secrets Manager.
// EndpointURL is used to point to LocalStack; leave empty for real AWS.
func NewSecretsManagerProvider(ctx context.Context, opts SecretsManagerOptions) (Provider, error) {
	region := opts.Region
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	if opts.EndpointURL != "" && opts.AccessKeyID != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(
			opts.AccessKeyID,
			opts.SecretAccessKey,
			"",
		)
	}

	clientOpts := []func(*secretsmanager.Options){}
	if opts.EndpointURL != "" {
		clientOpts = append(clientOpts, func(o *secretsmanager.Options) {
			o.BaseEndpoint = aws.String(opts.EndpointURL)
		})
	}
	client := secretsmanager.NewFromConfig(cfg, clientOpts...)

	cacheOpts := []func(*secretcache.Cache){
		func(c *secretcache.Cache) { c.Client = client },
	}
	if opts.CacheTTL > 0 {
		cacheOpts = append(cacheOpts, func(c *secretcache.Cache) {
			c.CacheConfig.CacheItemTTL = opts.CacheTTL.Nanoseconds()
		})
	}

	cache, err := secretcache.New(cacheOpts...)
	if err != nil {
		return nil, fmt.Errorf("create secrets cache: %w", err)
	}

	return &SecretsManagerProvider{
		cache:  cache,
		secret: opts.SecretID,
	}, nil
}

// GetDBCredentials retrieves the secret string and parses it as DBCredentials.
func (p *SecretsManagerProvider) GetDBCredentials(ctx context.Context) (*DBCredentials, error) {
	raw, err := p.cache.GetSecretStringWithContext(ctx, p.secret)
	if err != nil {
		return nil, fmt.Errorf("get secret value: %w", err)
	}

	if raw == "" {
		return nil, fmt.Errorf("secret value is empty")
	}

	var creds DBCredentials
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return nil, fmt.Errorf("parse secret json: %w", err)
	}

	return &creds, nil
}

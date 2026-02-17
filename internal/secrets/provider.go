package secrets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Provider returns database credentials from an external source (e.g. AWS Secrets Manager).
type Provider interface {
	GetDBCredentials(ctx context.Context) (*DBCredentials, error)
}

// SecretsManagerProvider loads DBCredentials from AWS Secrets Manager (LocalStack or AWS).
type SecretsManagerProvider struct {
	client *secretsmanager.Client
	secret string
}

// SecretsManagerOptions configures the Secrets Manager provider.
type SecretsManagerOptions struct {
	EndpointURL     string
	Region          string
	SecretID        string
	AccessKeyID     string
	SecretAccessKey string
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
	return &SecretsManagerProvider{client: client, secret: opts.SecretID}, nil
}

// GetDBCredentials retrieves the secret string and parses it as DBCredentials.
func (p *SecretsManagerProvider) GetDBCredentials(ctx context.Context) (*DBCredentials, error) {
	out, err := p.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(p.secret),
	})
	if err != nil {
		return nil, fmt.Errorf("get secret value: %w", err)
	}

	var raw string
	if out.SecretString != nil {
		raw = *out.SecretString
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

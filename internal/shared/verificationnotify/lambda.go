package verificationnotify

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaNotifierOptions configures async Lambda invocation for verification emails.
type LambdaNotifierOptions struct {
	EndpointURL     string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	FunctionName    string // function name, partial ARN, or full ARN
}

type lambdaAPI interface {
	Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

// LambdaNotifier invokes a Lambda function with InvocationType Event (asynchronous).
type LambdaNotifier struct {
	client       lambdaAPI
	functionName string
}

type verificationPayload struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Link  string `json:"link"`
}

type passwordResetPayload struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Link      string `json:"link"`
	ExpiresIn string `json:"expiresIn"`
}

// NewLambdaNotifier builds a Notifier that invokes the given function asynchronously.
func NewLambdaNotifier(ctx context.Context, opts LambdaNotifierOptions) (*LambdaNotifier, error) {
	fn := strings.TrimSpace(opts.FunctionName)
	if fn == "" {
		return nil, fmt.Errorf("lambda function name or ARN is required")
	}
	region := strings.TrimSpace(opts.Region)
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config for lambda: %w", err)
	}

	if opts.AccessKeyID != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(
			opts.AccessKeyID,
			opts.SecretAccessKey,
			"",
		)
	}

	clientOpts := []func(*lambda.Options){}
	if opts.EndpointURL != "" {
		clientOpts = append(clientOpts, func(o *lambda.Options) {
			o.BaseEndpoint = aws.String(opts.EndpointURL)
		})
	}

	client := lambda.NewFromConfig(cfg, clientOpts...)

	return &LambdaNotifier{
		client:       client,
		functionName: fn,
	}, nil
}

// NotifyVerificationEmail implements Notifier.
func (n *LambdaNotifier) NotifyVerificationEmail(ctx context.Context, toEmail, name, link string) error {
	to := strings.TrimSpace(toEmail)
	if to == "" {
		return fmt.Errorf("recipient email is required")
	}
	if strings.TrimSpace(link) == "" {
		return fmt.Errorf("verification link is required")
	}

	body, err := json.Marshal(verificationPayload{
		Email: to,
		Name:  name,
		Link:  link,
	})
	if err != nil {
		return fmt.Errorf("marshal verification payload: %w", err)
	}

	out, err := n.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(n.functionName),
		InvocationType: types.InvocationTypeEvent,
		Payload:        body,
	})
	if err != nil {
		return fmt.Errorf("lambda invoke: %w", err)
	}

	// Event invocations return 202; some local emulators may return 200.
	if out.StatusCode != 202 && out.StatusCode != 200 {
		return fmt.Errorf("lambda invoke unexpected status %d", out.StatusCode)
	}
	return nil
}

// NotifyPasswordResetEmail invokes the configured Lambda asynchronously with email, name, reset link, and expiresIn for SES template data.
func (n *LambdaNotifier) NotifyPasswordResetEmail(ctx context.Context, toEmail, name, link, expiresIn string) error {
	to := strings.TrimSpace(toEmail)
	if to == "" {
		return fmt.Errorf("recipient email is required")
	}
	if strings.TrimSpace(link) == "" {
		return fmt.Errorf("password reset link is required")
	}
	if strings.TrimSpace(expiresIn) == "" {
		return fmt.Errorf("expiresIn is required")
	}

	body, err := json.Marshal(passwordResetPayload{
		Email:     to,
		Name:      strings.TrimSpace(name),
		Link:      link,
		ExpiresIn: expiresIn,
	})
	if err != nil {
		return fmt.Errorf("marshal password reset payload: %w", err)
	}

	out, err := n.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(n.functionName),
		InvocationType: types.InvocationTypeEvent,
		Payload:        body,
	})
	if err != nil {
		return fmt.Errorf("lambda invoke: %w", err)
	}

	if out.StatusCode != 202 && out.StatusCode != 200 {
		return fmt.Errorf("lambda invoke unexpected status %d", out.StatusCode)
	}
	return nil
}

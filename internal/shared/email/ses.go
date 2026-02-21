package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// sesAPI is the subset of the SES v2 client used by SESSender, defined as an
// interface so it can be stubbed in tests.
type sesAPI interface {
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

// SESSender implements Sender using AWS SES v2.
type SESSender struct {
	client sesAPI
	from   string
	appURL string
}

// SESSenderOptions configures the SES sender.
type SESSenderOptions struct {
	EndpointURL     string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	FromAddress     string
	AppURL          string
}

// NewSESSender creates a Sender backed by AWS SES v2.
func NewSESSender(ctx context.Context, opts SESSenderOptions) (Sender, error) {
	region := opts.Region
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config for ses: %w", err)
	}

	if opts.AccessKeyID != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(
			opts.AccessKeyID,
			opts.SecretAccessKey,
			"",
		)
	}

	clientOpts := []func(*sesv2.Options){}
	if opts.EndpointURL != "" {
		clientOpts = append(clientOpts, func(o *sesv2.Options) {
			o.BaseEndpoint = aws.String(opts.EndpointURL)
		})
	}

	client := sesv2.NewFromConfig(cfg, clientOpts...)

	from := strings.TrimSpace(opts.FromAddress)
	if from == "" {
		return nil, fmt.Errorf("SES FromAddress is required and cannot be empty")
	}

	return &SESSender{
		client: client,
		from:   from,
		appURL: strings.TrimSuffix(strings.TrimSpace(opts.AppURL), "/"),
	}, nil
}

// SendVerificationEmail sends an account verification email to the given address.
func (s *SESSender) SendVerificationEmail(toAddress, toName, token string) error {
	to := strings.TrimSpace(toAddress)
	if to == "" {
		return fmt.Errorf("recipient email address is required and cannot be empty")
	}

	link := fmt.Sprintf("%s/auth/verify-email?token=%s", s.appURL, token)

	subject := "Verify your Cloudflax account"
	htmlBody := buildVerificationHTML(toName, link)
	textBody := buildVerificationText(toName, link)

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.from),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String(subject),
					Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data:    aws.String(htmlBody),
						Charset: aws.String("UTF-8"),
					},
					Text: &types.Content{
						Data:    aws.String(textBody),
						Charset: aws.String("UTF-8"),
					},
				},
			},
		},
	}

	if _, err := s.client.SendEmail(context.Background(), input); err != nil {
		return fmt.Errorf("ses send verification email: %w", err)
	}

	return nil
}

func buildVerificationHTML(name, link string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Verify your account</title></head>
<body style="font-family:sans-serif;color:#1a1a1a;max-width:600px;margin:0 auto;padding:32px 24px;">
  <h2 style="margin-bottom:8px;">Welcome to Cloudflax, %s!</h2>
  <p>Please verify your email address to activate your account.</p>
  <p style="margin:32px 0;">
    <a href="%s"
       style="background:#2563eb;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:600;">
      Verify email address
    </a>
  </p>
  <p style="color:#6b7280;font-size:14px;">
    Or copy and paste this link into your browser:<br>
    <a href="%s" style="color:#2563eb;">%s</a>
  </p>
  <p style="color:#6b7280;font-size:14px;">This link expires in 24 hours.</p>
</body>
</html>`, name, link, link, link)
}

func buildVerificationText(name, link string) string {
	return fmt.Sprintf(
		"Welcome to Cloudflax, %s!\n\nPlease verify your email address by visiting the link below:\n\n%s\n\nThis link expires in 24 hours.",
		name, link,
	)
}

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

// En: sesAPI is the subset of the SES v2 client used by SESSender, defined as an interface.
// Es: interface que define los métodos que se usarán para enviar emails usando SES v2.
type sesAPI interface {
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

// En: SESSender is the implementation of the TemplatedSender interface using AWS SES v2.
// Es: Implementación de la interfaz TemplatedSender usando AWS SES v2.
type SESSender struct {
	client sesAPI
	from   string
}

// En: SESSenderOptions configures the SES sender.
// Es: Opciones para configurar el sender de SES.
type SESSenderOptions struct {
	EndpointURL     string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	FromAddress     string
}

// En: NewSESSender creates a TemplatedSender backed by AWS SES v2.
// Es: Crea un TemplatedSender usando AWS SES v2.
func NewSESSender(ctx context.Context, opts SESSenderOptions) (TemplatedSender, error) {
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
	}, nil
}

// En: SendTemplatedEmail sends an email using an SES template. templateData must be a JSON object string with the variables expected by the template (e.g. {"name":"...","link":"..."}).
// Es: Envía un email usando un template de SES. templateData debe ser una cadena JSON con las variables esperadas por el template (ej. {"name":"...","link":"..."}).
func (s *SESSender) SendTemplatedEmail(toAddress, templateName, templateData string) error {
	to := strings.TrimSpace(toAddress)
	if to == "" {
		return fmt.Errorf("recipient email address is required and cannot be empty")
	}
	if strings.TrimSpace(templateName) == "" {
		return fmt.Errorf("template name is required and cannot be empty")
	}

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.from),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Template: &types.Template{
				TemplateName: aws.String(templateName),
				TemplateData: aws.String(templateData),
			},
		},
	}

	if _, err := s.client.SendEmail(context.Background(), input); err != nil {
		return fmt.Errorf("ses send templated email: %w", err)
	}

	return nil
}

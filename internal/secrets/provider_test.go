package secrets

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type stubSecretsClient struct {
	calls     int
	responses []struct {
		out *secretsmanager.GetSecretValueOutput
		err error
	}
}

func (s *stubSecretsClient) GetSecretValue(ctx context.Context, in *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	_ = ctx
	_ = in
	_ = optFns
	s.calls++
	if len(s.responses) == 0 {
		return &secretsmanager.GetSecretValueOutput{}, nil
	}
	idx := s.calls - 1
	if idx >= len(s.responses) {
		idx = len(s.responses) - 1
	}
	r := s.responses[idx]
	return r.out, r.err
}

func TestSecretsManagerProvider_Success(t *testing.T) {
	t.Helper()

	raw := `{"dbname":"cloudflax","host":"h","password":"p","port":4510,"username":"u"}`
	client := &stubSecretsClient{
		responses: []struct {
			out *secretsmanager.GetSecretValueOutput
			err error
		}{
			{out: &secretsmanager.GetSecretValueOutput{SecretString: aws.String(raw)}, err: nil},
		},
	}

	p := &SecretsManagerProvider{client: client, secret: "any"}

	creds, err := p.GetDBCredentials(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.DBName() != "cloudflax" || creds.Username() != "u" || creds.Password() != "p" {
		t.Fatalf("unexpected creds: %+v", creds)
	}
	if client.calls != 1 {
		t.Fatalf("expected one call to GetSecretValue, got %d", client.calls)
	}
}

func TestSecretsManagerProvider_EmptySecret(t *testing.T) {
	t.Helper()

	client := &stubSecretsClient{
		responses: []struct {
			out *secretsmanager.GetSecretValueOutput
			err error
		}{
			{out: &secretsmanager.GetSecretValueOutput{SecretString: nil}, err: nil},
		},
	}

	p := &SecretsManagerProvider{client: client, secret: "any"}

	_, err := p.GetDBCredentials(context.Background())
	if err == nil {
		t.Fatalf("expected error for empty secret, got nil")
	}
}

func TestSecretsManagerProvider_AWSErrorWrapped(t *testing.T) {
	t.Helper()

	client := &stubSecretsClient{
		responses: []struct {
			out *secretsmanager.GetSecretValueOutput
			err error
		}{
			{out: nil, err: errors.New("boom")},
		},
	}

	p := &SecretsManagerProvider{client: client, secret: "any"}

	_, err := p.GetDBCredentials(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if client.calls != 1 {
		t.Fatalf("expected one call to GetSecretValue, got %d", client.calls)
	}
}


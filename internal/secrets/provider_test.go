package secrets

import (
	"context"
	"errors"
	"testing"
)

type stubSecretCache struct {
	calls     int
	responses []struct {
		value string
		err   error
	}
}

func (s *stubSecretCache) GetSecretStringWithContext(ctx context.Context, secretId string) (string, error) {
	_ = ctx
	_ = secretId

	s.calls++
	if len(s.responses) == 0 {
		return "", nil
	}
	idx := s.calls - 1
	if idx >= len(s.responses) {
		idx = len(s.responses) - 1
	}
	r := s.responses[idx]
	return r.value, r.err
}

func TestSecretsManagerProvider_Success(t *testing.T) {
	t.Helper()

	raw := `{"dbname":"cloudflax","host":"h","password":"p","port":4510,"username":"u"}`
	cache := &stubSecretCache{
		responses: []struct {
			value string
			err   error
		}{
			{value: raw, err: nil},
		},
	}

	p := &SecretsManagerProvider{cache: cache, secret: "any"}

	creds, err := p.GetDBCredentials(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.DBName() != "cloudflax" || creds.Username() != "u" || creds.Password() != "p" {
		t.Fatalf("unexpected creds: %+v", creds)
	}
	if cache.calls != 1 {
		t.Fatalf("expected one call to GetSecretStringWithContext, got %d", cache.calls)
	}
}

func TestSecretsManagerProvider_EmptySecret(t *testing.T) {
	t.Helper()

	cache := &stubSecretCache{
		responses: []struct {
			value string
			err   error
		}{
			{value: "", err: nil},
		},
	}

	p := &SecretsManagerProvider{cache: cache, secret: "any"}

	_, err := p.GetDBCredentials(context.Background())
	if err == nil {
		t.Fatalf("expected error for empty secret, got nil")
	}
}

func TestSecretsManagerProvider_AWSErrorWrapped(t *testing.T) {
	t.Helper()

	cache := &stubSecretCache{
		responses: []struct {
			value string
			err   error
		}{
			{value: "", err: errors.New("boom")},
		},
	}

	p := &SecretsManagerProvider{cache: cache, secret: "any"}

	_, err := p.GetDBCredentials(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if cache.calls != 1 {
		t.Fatalf("expected one call to GetSecretStringWithContext, got %d", cache.calls)
	}
}


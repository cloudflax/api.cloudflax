package secrets

import (
	"context"
	"testing"
	"time"
)

type stubProvider struct {
	calls int
	// sequence of results; if err is non-nil, creds is ignored
	results []struct {
		creds *DBCredentials
		err   error
	}
}

func (s *stubProvider) GetDBCredentials(ctx context.Context) (*DBCredentials, error) {
	_ = ctx
	s.calls++
	if len(s.results) == 0 {
		return &DBCredentials{}, nil
	}
	idx := s.calls - 1
	if idx >= len(s.results) {
		idx = len(s.results) - 1
	}
	r := s.results[idx]
	if r.err != nil {
		return nil, r.err
	}
	return r.creds, nil
}

func TestCachingProvider_UsesCacheWithinTTL(t *testing.T) {
	t.Helper()

	stub := &stubProvider{
		results: []struct {
			creds *DBCredentials
			err   error
		}{
			{creds: &DBCredentials{dbName: "cloudflax", host: "h", username: "u", password: "p", port: 5432}},
		},
	}

	cache := NewCachingProvider(stub, time.Minute)

	ctx := context.Background()
	_, err := cache.GetDBCredentials(ctx)
	if err != nil {
		t.Fatalf("first call returned error: %v", err)
	}
	_, err = cache.GetDBCredentials(ctx)
	if err != nil {
		t.Fatalf("second call returned error: %v", err)
	}

	if stub.calls != 1 {
		t.Fatalf("expected inner provider to be called once, got %d", stub.calls)
	}
}

func TestCachingProvider_NoCacheWhenTTLZero(t *testing.T) {
	t.Helper()

	stub := &stubProvider{
		results: []struct {
			creds *DBCredentials
			err   error
		}{
			{creds: &DBCredentials{dbName: "cloudflax"}},
		},
	}

	cache := NewCachingProvider(stub, 0)
	ctx := context.Background()

	_, _ = cache.GetDBCredentials(ctx)
	_, _ = cache.GetDBCredentials(ctx)

	if stub.calls != 2 {
		t.Fatalf("expected inner provider to be called twice with ttl=0, got %d", stub.calls)
	}
}

func TestCachingProvider_RetryOnErrorOnce(t *testing.T) {
	t.Helper()

	expected := &DBCredentials{dbName: "cloudflax"}
	stub := &stubProvider{
		results: []struct {
			creds *DBCredentials
			err   error
		}{
			{creds: nil, err: context.DeadlineExceeded}, // first call fails
			{creds: expected, err: nil},                 // second call succeeds
		},
	}

	cache := NewCachingProvider(stub, time.Minute)
	ctx := context.Background()

	creds, err := cache.GetDBCredentials(ctx)
	if err != nil {
		t.Fatalf("expected retry to succeed, got error: %v", err)
	}
	if creds.DBName() != expected.DBName() {
		t.Fatalf("unexpected db name, got %q want %q", creds.DBName(), expected.DBName())
	}
	if stub.calls != 2 {
		t.Fatalf("expected two calls to inner provider (error + retry), got %d", stub.calls)
	}
}


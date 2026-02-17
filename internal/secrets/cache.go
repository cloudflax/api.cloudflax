package secrets

import (
	"context"
	"sync"
	"time"
)

// cachingProvider wraps a Provider and caches the result for a configurable TTL.
// This avoids hitting AWS Secrets Manager (o LocalStack) en cada llamada.
type cachingProvider struct {
	inner Provider
	ttl   time.Duration

	mu      sync.RWMutex
	cached  *DBCredentials
	expires time.Time
}

// NewCachingProvider returns a Provider that caches the DB credentials for the
// given TTL. If ttl <= 0, it returns inner sin caché.
func NewCachingProvider(inner Provider, ttl time.Duration) Provider {
	if ttl <= 0 {
		return inner
	}
	return &cachingProvider{
		inner: inner,
		ttl:   ttl,
	}
}

// GetDBCredentials returns cached credentials when válidas; si el caché expira,
// vuelve a llamar al proveedor interno y refresca el valor.
func (c *cachingProvider) GetDBCredentials(ctx context.Context) (*DBCredentials, error) {
	now := time.Now()

	// Lectura rápida con RLock
	c.mu.RLock()
	if c.cached != nil && now.Before(c.expires) {
		credsCopy := *c.cached
		c.mu.RUnlock()
		return &credsCopy, nil
	}
	c.mu.RUnlock()

	// Necesitamos refrescar el caché
	c.mu.Lock()
	defer c.mu.Unlock()

	// Revalidar por si otro goroutine ya refrescó
	now = time.Now()
	if c.cached != nil && now.Before(c.expires) {
		credsCopy := *c.cached
		return &credsCopy, nil
	}

	creds, err := c.inner.GetDBCredentials(ctx)
	if err != nil {
		return nil, err
	}

	// Guardar copia para evitar aliasing
	credsCopy := *creds
	c.cached = &credsCopy
	c.expires = now.Add(c.ttl)

	return creds, nil
}


package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveSlowQueryThresholdMS(t *testing.T) {
	tests := []struct {
		name     string
		envSlow  string
		appEnv   string
		wantMS   int
	}{
		{
			name:    "explicit override",
			envSlow: "750",
			appEnv:  "production",
			wantMS:  750,
		},
		{
			name:    "development default when unset",
			envSlow: "",
			appEnv:  "development",
			wantMS:  500,
		},
		{
			name:    "production default when unset",
			envSlow: "",
			appEnv:  "production",
			wantMS:  200,
		},
		{
			name:    "empty APP_ENV defaults to non-development",
			envSlow: "",
			appEnv:  "",
			wantMS:  200,
		},
		{
			name:    "zero env falls back to APP_ENV default",
			envSlow: "0",
			appEnv:  "development",
			wantMS:  500,
		},
		{
			name:    "negative env falls back to APP_ENV default",
			envSlow: "-1",
			appEnv:  "development",
			wantMS:  500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DB_SLOW_QUERY_THRESHOLD_MS", tt.envSlow)
			t.Setenv("APP_ENV", tt.appEnv)
			assert.Equal(t, tt.wantMS, resolveSlowQueryThresholdMS())
		})
	}
}

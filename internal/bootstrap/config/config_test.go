package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnvBool(t *testing.T) {
	t.Run("empty uses default", func(t *testing.T) {
		t.Setenv("TEST_ENV_BOOL_X", "")
		assert.True(t, envBool("TEST_ENV_BOOL_X", true))
		assert.False(t, envBool("TEST_ENV_BOOL_X", false))
	})
	t.Run("truthy", func(t *testing.T) {
		for _, v := range []string{"1", "true", "TRUE", "yes", "On"} {
			t.Setenv("TEST_ENV_BOOL_Y", v)
			assert.True(t, envBool("TEST_ENV_BOOL_Y", false), v)
		}
	})
	t.Run("falsey", func(t *testing.T) {
		for _, v := range []string{"0", "false", "no", "off"} {
			t.Setenv("TEST_ENV_BOOL_Z", v)
			assert.False(t, envBool("TEST_ENV_BOOL_Z", true), v)
		}
	})
}

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

func TestConfigValidateProductionVsAuthDevEndpoints(t *testing.T) {
	t.Run("production with dev auth endpoints fails", func(t *testing.T) {
		cfg := &Config{
			Port:                   "3000",
			AppEnv:                 "production",
			JWTSecret:              "x",
			DBHost:                 "h",
			DBUser:                 "u",
			DBName:                 "n",
			EnableAuthDevEndpoints: true,
			JWTAccessTokenDuration: 15 * time.Minute,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ENABLE_AUTH_DEV_ENDPOINTS")
	})
	t.Run("production with dev endpoints off is ok", func(t *testing.T) {
		cfg := &Config{
			Port:                   "3000",
			AppEnv:                 "production",
			JWTSecret:              "x",
			DBHost:                 "h",
			DBUser:                 "u",
			DBName:                 "n",
			EnableAuthDevEndpoints: false,
			JWTAccessTokenDuration: 15 * time.Minute,
		}
		assert.NoError(t, cfg.Validate())
	})
	t.Run("staging with dev endpoints allowed", func(t *testing.T) {
		cfg := &Config{
			Port:                   "3000",
			AppEnv:                 "staging",
			JWTSecret:              "x",
			DBHost:                 "h",
			DBUser:                 "u",
			DBName:                 "n",
			EnableAuthDevEndpoints: true,
			JWTAccessTokenDuration: 15 * time.Minute,
		}
		assert.NoError(t, cfg.Validate())
	})
}

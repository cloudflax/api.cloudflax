package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cloudflax/api.cloudflax/internal/auth"
	"github.com/cloudflax/api.cloudflax/internal/bootstrap/config"
)

const throttleGuardInitFailedEvent = "api_throttle_guard_init_failed"

// attachAuthThrottleGuards wires Dynamo-backed auth throttles onto the handler.
// When API_THROTTLE_TABLE_NAME is empty, all guards are skipped (fail-open: no IP/email throttles).
// When the table name is set but a client cannot be created, behavior depends on API_THROTTLE_STRICT_INIT.
func attachAuthThrottleGuards(handler *auth.Handler, cfg *config.Config) (*auth.Handler, error) {
	table := strings.TrimSpace(cfg.APIThrottleTableName)
	opts := auth.DynamoResendVerificationGuardOptions{
		TableName:       cfg.APIThrottleTableName,
		EndpointURL:     cfg.AWSEndpointURL,
		Region:          cfg.AWSRegion,
		Profile:         cfg.AWSProfile,
		AccessKeyID:     cfg.AWSAccessKeyID,
		SecretAccessKey: cfg.AWSSecretAccessKey,
	}
	ipOpts := auth.DynamoIPThrottleGuardOptions{
		TableName:       cfg.APIThrottleTableName,
		EndpointURL:     cfg.AWSEndpointURL,
		Region:          cfg.AWSRegion,
		Profile:         cfg.AWSProfile,
		AccessKeyID:     cfg.AWSAccessKeyID,
		SecretAccessKey: cfg.AWSSecretAccessKey,
	}

	ctx := context.Background()

	resendGuard, err := auth.NewDynamoResendVerificationGuard(ctx, opts)
	if err != nil {
		if cfg.APIThrottleStrictInit && table != "" {
			return nil, fmt.Errorf("resend verification throttle init: %w", err)
		}
		slog.Warn("failed to initialise resend verification guard",
			"event", throttleGuardInitFailedEvent,
			"component", "resend_verification",
			"error", err,
		)
	} else if resendGuard != nil {
		handler = handler.WithResendVerificationGuard(resendGuard)
	}

	forgotGuard, err := auth.NewDynamoForgotPasswordGuard(ctx, auth.DynamoForgotPasswordGuardOptions{
		TableName:       cfg.APIThrottleTableName,
		EndpointURL:     cfg.AWSEndpointURL,
		Region:          cfg.AWSRegion,
		Profile:         cfg.AWSProfile,
		AccessKeyID:     cfg.AWSAccessKeyID,
		SecretAccessKey: cfg.AWSSecretAccessKey,
	})
	if err != nil {
		if cfg.APIThrottleStrictInit && table != "" {
			return nil, fmt.Errorf("forgot-password throttle init: %w", err)
		}
		slog.Warn("failed to initialise forgot-password guard",
			"event", throttleGuardInitFailedEvent,
			"component", "forgot_password",
			"error", err,
		)
	} else if forgotGuard != nil {
		handler = handler.WithForgotPasswordGuard(forgotGuard)
	}

	loginThrottle, err := auth.NewDynamoLoginIPThrottleGuard(ctx, ipOpts)
	if err != nil {
		if cfg.APIThrottleStrictInit && table != "" {
			return nil, fmt.Errorf("login IP throttle init: %w", err)
		}
		slog.Warn("failed to initialise login IP throttle guard",
			"event", throttleGuardInitFailedEvent,
			"component", "login_ip",
			"error", err,
		)
	} else if loginThrottle != nil {
		handler = handler.WithLoginIPThrottleGuard(loginThrottle)
	}

	refreshThrottle, err := auth.NewDynamoRefreshIPThrottleGuard(ctx, ipOpts)
	if err != nil {
		if cfg.APIThrottleStrictInit && table != "" {
			return nil, fmt.Errorf("refresh IP throttle init: %w", err)
		}
		slog.Warn("failed to initialise refresh IP throttle guard",
			"event", throttleGuardInitFailedEvent,
			"component", "refresh_ip",
			"error", err,
		)
	} else if refreshThrottle != nil {
		handler = handler.WithRefreshIPThrottleGuard(refreshThrottle)
	}

	return handler, nil
}

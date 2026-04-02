package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// En: TestEvaluateResendStateFirstSendAllows verifies first resend attempt is allowed.
// Es: TestEvaluateResendStateFirstSendAllows verifica que el primer reenvio se permite.
func TestEvaluateResendStateFirstSendAllows(test *testing.T) {
	now := int64(1_700_000_000)

	next, err := evaluateResendState(&resendGuardState{}, now)
	require.NoError(test, err)
	require.NotNil(test, next)

	assert.Equal(test, int64(1), next.Count)
	assert.Equal(test, now+resendCooldownSeconds, next.NextAllowedAt)
	assert.Equal(test, int64(0), next.LockUntil)
}

// En: TestEvaluateResendStateCooldownBlocks verifies resend is blocked inside cooldown.
// Es: TestEvaluateResendStateCooldownBlocks verifica bloqueo de reenvio en cooldown.
func TestEvaluateResendStateCooldownBlocks(test *testing.T) {
	now := int64(1_700_000_000)
	current := &resendGuardState{
		Exists:        true,
		Count:         1,
		NextAllowedAt: now + resendCooldownSeconds,
	}

	next, err := evaluateResendState(current, now+30)
	require.Error(test, err)
	assert.Nil(test, next)
	assert.ErrorIs(test, err, ErrResendVerificationRateLimited)
}

// En: TestEvaluateResendStateThirdSendLocks verifies third resend applies lock period.
// Es: TestEvaluateResendStateThirdSendLocks verifica que tercer reenvio activa bloqueo.
func TestEvaluateResendStateThirdSendLocks(test *testing.T) {
	now := int64(1_700_000_000)
	current := &resendGuardState{
		Exists:        true,
		Version:       2,
		Count:         2,
		WindowStart:   now - 600,
		NextAllowedAt: now,
		CreatedAt:     now - 600,
		OriginalPK:    "THROTTLE#RESEND_VERIFICATION#EMAIL#hash",
		OriginalSK:    resendStateSK,
	}

	next, err := evaluateResendState(current, now)
	require.NoError(test, err)
	require.NotNil(test, next)

	assert.Equal(test, int64(3), next.Count)
	assert.Equal(test, now+resendLockSeconds, next.LockUntil)
}

// En: TestEvaluateResendStateLockedBlocks verifies requests are blocked during lock.
// Es: TestEvaluateResendStateLockedBlocks verifica bloqueo de solicitudes durante lock.
func TestEvaluateResendStateLockedBlocks(test *testing.T) {
	now := int64(1_700_000_000)
	current := &resendGuardState{
		Exists:     true,
		Count:      3,
		LockUntil:  now + resendLockSeconds,
		OriginalPK: "THROTTLE#RESEND_VERIFICATION#EMAIL#hash",
		OriginalSK: resendStateSK,
	}

	next, err := evaluateResendState(current, now+60)
	require.Error(test, err)
	assert.Nil(test, next)
	assert.ErrorIs(test, err, ErrResendVerificationRateLimited)

	var limitErr *ResendVerificationRateLimitError
	require.True(test, errors.As(err, &limitErr))
	assert.Greater(test, limitErr.RetryAfter, time.Duration(0))
}

// En: TestEvaluateResendStateAfterLockResets verifies sending resumes after lock expiry.
// Es: TestEvaluateResendStateAfterLockResets verifica que reenvio vuelve tras expirar lock.
func TestEvaluateResendStateAfterLockResets(test *testing.T) {
	now := int64(1_700_000_000)
	current := &resendGuardState{
		Exists:      true,
		Version:     7,
		Count:       3,
		LockUntil:   now - 1,
		CreatedAt:   now - 1000,
		OriginalPK:  "THROTTLE#RESEND_VERIFICATION#EMAIL#hash",
		OriginalSK:  resendStateSK,
		WindowStart: now - 1000,
	}

	next, err := evaluateResendState(current, now)
	require.NoError(test, err)
	require.NotNil(test, next)

	assert.Equal(test, int64(1), next.Count)
	assert.Equal(test, now+resendCooldownSeconds, next.NextAllowedAt)
	assert.Equal(test, int64(0), next.LockUntil)
}

package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// En: TestEvaluateIPWindowFirstAttempt verifies the first request in a window is allowed.
// Es: TestEvaluateIPWindowFirstAttempt verifica la primera solicitud en una ventana.
func TestEvaluateIPWindowFirstAttempt(test *testing.T) {
	now := int64(1_700_000_000)
	next, limitErr := evaluateIPWindowState(&resendGuardState{}, now, 30, 600, 1800, 86400)
	require.Nil(test, limitErr)
	require.NotNil(test, next)
	assert.Equal(test, int64(1), next.Count)
	assert.Equal(test, now, next.WindowStart)
	assert.Equal(test, int64(0), next.LockUntil)
}

// En: TestEvaluateIPWindowIncrementsInSameWindow keeps window start stable.
// Es: TestEvaluateIPWindowIncrementsInSameWindow mantiene window_start estable.
func TestEvaluateIPWindowIncrementsInSameWindow(test *testing.T) {
	now := int64(1_700_000_000)
	current := &resendGuardState{
		Exists:        true,
		Version:       1,
		Count:         5,
		WindowStart:   now - 120,
		OriginalPK:    "THROTTLE#LOGIN#IP#ab",
		OriginalSK:    resendStateSK,
	}

	next, limitErr := evaluateIPWindowState(current, now, 30, 600, 1800, 86400)
	require.Nil(test, limitErr)
	require.NotNil(test, next)
	assert.Equal(test, int64(6), next.Count)
	assert.Equal(test, current.WindowStart, next.WindowStart)
}

// En: TestEvaluateIPWindowBlocksWhenOverMax emits a lock and rate-limit metadata.
// Es: TestEvaluateIPWindowBlocksWhenOverMax emite bloqueo y metadatos de limite.
func TestEvaluateIPWindowBlocksWhenOverMax(test *testing.T) {
	now := int64(1_700_000_000)
	current := &resendGuardState{
		Exists:      true,
		Version:     10,
		Count:       30,
		WindowStart: now - 60,
		OriginalPK:  "THROTTLE#LOGIN#IP#ab",
		OriginalSK:  resendStateSK,
	}

	next, limitErr := evaluateIPWindowState(current, now, 30, 600, 1800, 86400)
	require.NotNil(test, limitErr)
	require.ErrorIs(test, limitErr, ErrIPThrottleRateLimited)
	require.NotNil(test, next)
	assert.Equal(test, int64(31), next.Count)
	assert.Equal(test, now+1800, next.LockUntil)
	assert.Equal(test, 1800*time.Second, limitErr.RetryAfter)
}

// En: TestEvaluateIPWindowActiveLock blocks until lock expires.
// Es: TestEvaluateIPWindowActiveLock bloquea hasta que expire el lock.
func TestEvaluateIPWindowActiveLock(test *testing.T) {
	now := int64(1_700_000_000)
	current := &resendGuardState{
		Exists:     true,
		LockUntil:  now + 300,
		OriginalPK: "THROTTLE#LOGIN#IP#ab",
		OriginalSK: resendStateSK,
	}

	next, limitErr := evaluateIPWindowState(current, now, 30, 600, 1800, 86400)
	assert.Nil(test, next)
	require.NotNil(test, limitErr)
	assert.Equal(test, 300*time.Second, limitErr.RetryAfter)
}

// En: TestEvaluateIPWindowResetsAfterWindowElapsed starts a new window.
// Es: TestEvaluateIPWindowResetsAfterWindowElapsed inicia ventana nueva.
func TestEvaluateIPWindowResetsAfterWindowElapsed(test *testing.T) {
	now := int64(1_700_000_000)
	current := &resendGuardState{
		Exists:        true,
		Version:       3,
		Count:         30,
		WindowStart:   now - 700,
		OriginalPK:    "THROTTLE#LOGIN#IP#ab",
		OriginalSK:    resendStateSK,
	}

	next, limitErr := evaluateIPWindowState(current, now, 30, 600, 1800, 86400)
	require.Nil(test, limitErr)
	require.NotNil(test, next)
	assert.Equal(test, int64(1), next.Count)
	assert.Equal(test, now, next.WindowStart)
}

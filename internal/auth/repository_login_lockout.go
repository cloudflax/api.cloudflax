package auth

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	loginCredentialLockoutMaxFailures = 5
	loginCredentialLockoutWindow      = 15 * time.Minute
	loginCredentialLockoutDuration    = 30 * time.Minute
)

// En: LoginCredentialLockRetryAfter returns remaining lock duration for normalizedEmail. Zero duration means not locked.
// Es: LoginCredentialLockRetryAfter devuelve el tiempo restante de bloqueo. Duración cero significa que no hay bloqueo.
func (repository *Repository) LoginCredentialLockRetryAfter(normalizedEmail string) (time.Duration, error) {
	if normalizedEmail == "" {
		return 0, nil
	}
	now := time.Now()
	var row LoginCredentialLockout
	err := repository.db.Where("email_normalized = ?", normalizedEmail).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("load login credential lockout: %w", err)
	}
	if row.LockedUntil != nil && row.LockedUntil.After(now) {
		return row.LockedUntil.Sub(now), nil
	}
	return 0, nil
}

// En: RecordFailedLoginCredentialAttempt increments failures for the email and applies lockout when limits are exceeded.
// Es: RecordFailedLoginCredentialAttempt incrementa fallos para el email y aplica bloqueo al superar limites.
func (repository *Repository) RecordFailedLoginCredentialAttempt(normalizedEmail string) error {
	if normalizedEmail == "" {
		return nil
	}
	now := time.Now()
	return repository.db.Transaction(func(tx *gorm.DB) error {
		var row LoginCredentialLockout
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("email_normalized = ?", normalizedEmail).
			First(&row).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&LoginCredentialLockout{
				EmailNormalized: normalizedEmail,
				FailedCount:     1,
				WindowStart:     now,
			}).Error
		}
		if err != nil {
			return fmt.Errorf("load login credential lockout: %w", err)
		}

		if row.LockedUntil != nil && row.LockedUntil.After(now) {
			return nil
		}

		if row.LockedUntil != nil && !row.LockedUntil.After(now) {
			row.LockedUntil = nil
			row.FailedCount = 1
			row.WindowStart = now
			return tx.Save(&row).Error
		}

		if now.Sub(row.WindowStart) > loginCredentialLockoutWindow {
			row.FailedCount = 1
			row.WindowStart = now
			row.LockedUntil = nil
			return tx.Save(&row).Error
		}

		row.FailedCount++
		if row.FailedCount >= loginCredentialLockoutMaxFailures {
			lockUntil := now.Add(loginCredentialLockoutDuration)
			row.LockedUntil = &lockUntil
		}
		return tx.Save(&row).Error
	})
}

// En: ClearLoginCredentialLockout removes lockout state after a successful credential check.
// Es: ClearLoginCredentialLockout elimina el estado de bloqueo tras una verificacion de credencial exitosa.
func (repository *Repository) ClearLoginCredentialLockout(normalizedEmail string) error {
	if normalizedEmail == "" {
		return nil
	}
	if err := repository.db.Where("email_normalized = ?", normalizedEmail).Delete(&LoginCredentialLockout{}).Error; err != nil {
		return fmt.Errorf("clear login credential lockout: %w", err)
	}
	return nil
}

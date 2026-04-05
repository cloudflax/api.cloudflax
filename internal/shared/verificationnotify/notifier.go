package verificationnotify

import "context"

// Notifier triggers delivery of the email verification message (e.g. async Lambda that sends via SES).
type Notifier interface {
	NotifyVerificationEmail(ctx context.Context, toEmail, name, link string) error
}

// NoopNotifier is a Notifier that does nothing.
type NoopNotifier struct{}

// NotifyVerificationEmail implements Notifier.
func (NoopNotifier) NotifyVerificationEmail(context.Context, string, string, string) error {
	return nil
}

// PasswordResetEmailNotifier triggers delivery of the forgot-password message (async Lambda → SES template).
type PasswordResetEmailNotifier interface {
	NotifyPasswordResetEmail(ctx context.Context, toEmail, name, link, expiresIn string) error
}

// NoopPasswordResetEmailNotifier is a PasswordResetEmailNotifier that does nothing.
type NoopPasswordResetEmailNotifier struct{}

// NotifyPasswordResetEmail implements PasswordResetEmailNotifier.
func (NoopPasswordResetEmailNotifier) NotifyPasswordResetEmail(context.Context, string, string, string, string) error {
	return nil
}

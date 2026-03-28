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

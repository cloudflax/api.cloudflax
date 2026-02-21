package email

// Sender sends transactional emails.
type Sender interface {
	SendVerificationEmail(toAddress, toName, token string) error
}

// NoopSender is a Sender that discards all messages. Useful in tests and
// environments where email delivery is not configured.
type NoopSender struct{}

func (n *NoopSender) SendVerificationEmail(_, _, _ string) error { return nil }

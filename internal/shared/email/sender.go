package email

// En: TemplatedSender sends emails using SES (or other) templates. Template name and template data are provided by the caller; this package does not define domain-specific emails (e.g. verification is handled by the auth module).
// Es: TemplatedSender envía correos usando plantillas SES (o otras). El nombre de la plantilla y los datos de la plantilla se proporcionan al llamador; este paquete no define correos específicos para el dominio (por ejemplo, la verificación de cuenta es manejada por el módulo de autenticación).
type TemplatedSender interface {
	SendTemplatedEmail(toAddress, templateName, templateData string) error
}

// En: NoopSender is a TemplatedSender that discards all messages. Useful in tests and environments where email delivery is not configured.
// Es: NoopSender es un TemplatedSender que descarta todos los mensajes. Útil en pruebas y entornos donde la entrega de correos no está configurada.
type NoopSender struct{}

// En: Sends an email using the given template name and data.
// Es: Envía un correo electrónico usando el nombre de la plantilla y los datos proporcionados.
func (n *NoopSender) SendTemplatedEmail(_, _, _ string) error { return nil }

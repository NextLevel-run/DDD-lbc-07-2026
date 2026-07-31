package mailer

// Mailer defines the interface for sending emails
type Mailer interface {
	SendSimpleEmail(toEmail, fromEmail, title, body string) error
	SendTemplatedEmail(toEmail, fromEmail, templateId string, templateData map[string]any) error
}

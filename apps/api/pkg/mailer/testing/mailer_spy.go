package testing

import "sync"

// SentSimpleEmail represents a simple email that was sent
type SentSimpleEmail struct {
	To    string
	From  string
	Title string
	Body  string
}

// SentTemplatedEmail represents a templated email that was sent
type SentTemplatedEmail struct {
	To           string
	From         string
	TemplateId   string
	TemplateData map[string]any
}

// MailerSpy is a test spy that captures sent emails for verification
type MailerSpy struct {
	sentSimpleEmails    []SentSimpleEmail
	sentTemplatedEmails []SentTemplatedEmail
	mu                  sync.Mutex
}

// NewMailerSpy creates a new MailerSpy instance
func NewMailerSpy() *MailerSpy {
	return &MailerSpy{
		sentSimpleEmails:    make([]SentSimpleEmail, 0),
		sentTemplatedEmails: make([]SentTemplatedEmail, 0),
	}
}

// SendSimpleEmail captures the simple email for testing
func (m *MailerSpy) SendSimpleEmail(toEmail, fromEmail, title, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentSimpleEmails = append(m.sentSimpleEmails, SentSimpleEmail{
		To:    toEmail,
		From:  fromEmail,
		Title: title,
		Body:  body,
	})

	return nil
}

// SendTemplatedEmail captures the templated email for testing
func (m *MailerSpy) SendTemplatedEmail(toEmail, fromEmail, templateId string, templateData map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentTemplatedEmails = append(m.sentTemplatedEmails, SentTemplatedEmail{
		To:           toEmail,
		From:         fromEmail,
		TemplateId:   templateId,
		TemplateData: templateData,
	})

	return nil
}

// GetSentSimpleEmails returns all captured simple emails
func (m *MailerSpy) GetSentSimpleEmails() []SentSimpleEmail {
	m.mu.Lock()
	defer m.mu.Unlock()

	emails := make([]SentSimpleEmail, len(m.sentSimpleEmails))
	copy(emails, m.sentSimpleEmails)
	return emails
}

// GetSentTemplatedEmails returns all captured templated emails
func (m *MailerSpy) GetSentTemplatedEmails() []SentTemplatedEmail {
	m.mu.Lock()
	defer m.mu.Unlock()

	emails := make([]SentTemplatedEmail, len(m.sentTemplatedEmails))
	copy(emails, m.sentTemplatedEmails)
	return emails
}

// Reset clears all captured emails
func (m *MailerSpy) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentSimpleEmails = make([]SentSimpleEmail, 0)
	m.sentTemplatedEmails = make([]SentTemplatedEmail, 0)
}

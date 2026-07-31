package mailer

import (
	"fmt"
)

// FakeMailer is a fake implementation of the Mailer interface
// that logs emails to the console instead of actually sending them
type FakeMailer struct{}

// NewFakeMailer creates a new FakeMailer instance
func NewFakeMailer() *FakeMailer {
	return &FakeMailer{}
}

// SendSimpleEmail logs the email to the console
func (m *FakeMailer) SendSimpleEmail(toEmail, fromEmail, title, body string) error {
	fmt.Printf("Send simple email from %s to %s: %s\n", fromEmail, toEmail, title)
	return nil
}

// SendTemplatedEmail logs the templated email to the console
func (m *FakeMailer) SendTemplatedEmail(toEmail, fromEmail, templateId string, templateData map[string]any) error {
	fmt.Printf("Send templated email from %s to %s using template %s\n", fromEmail, toEmail, templateId)
	return nil
}

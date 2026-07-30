package consumer_test

import (
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/consumer"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	mailertesting "ddd-second-hand-marketplace/pkg/mailer/testing"
)

func TestAdPublishedEmailConsumer_SendsEmailToSeller(t *testing.T) {
	spy := mailertesting.NewMailerSpy()
	handler := consumer.NewAdPublishedEmailConsumer(spy)

	event := &domain.ClassifiedAdPublishedEvent{
		AdID:         "ad-123",
		Title:        "Vélo hollandais",
		Category:     "consumer_goods",
		SellerEmail:  "seller@example.com",
		SellerPseudo: "SellerPseudo",
		PublishedAt:  time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC),
	}

	if err := handler(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sent := spy.GetSentSimpleEmails()
	if len(sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(sent))
	}

	email := sent[0]
	if email.To != "seller@example.com" {
		t.Errorf("expected recipient seller@example.com, got %s", email.To)
	}
	if email.From != "no-reply@marketplace.local" {
		t.Errorf("expected sender no-reply@marketplace.local, got %s", email.From)
	}
	if !strings.Contains(email.Title, "Vélo hollandais") {
		t.Errorf("expected title to mention ad title, got %s", email.Title)
	}
	if !strings.Contains(email.Body, "SellerPseudo") {
		t.Errorf("expected body to mention seller pseudo, got %s", email.Body)
	}
}

func TestAdPublishedEmailConsumer_IgnoresOtherEvents(t *testing.T) {
	spy := mailertesting.NewMailerSpy()
	handler := consumer.NewAdPublishedEmailConsumer(spy)

	event := &domain.ClassifiedAdDeletedEvent{AdID: "ad-123"}

	if err := handler(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.GetSentSimpleEmails()) != 0 {
		t.Errorf("expected no email sent for unrelated event")
	}
}

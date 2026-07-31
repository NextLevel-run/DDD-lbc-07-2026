package consumer_test

import (
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/consumer"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	mailertesting "ddd-second-hand-marketplace/pkg/mailer/testing"
)

func TestOfferEmailConsumer_SendsEmailToSeller(t *testing.T) {
	spy := mailertesting.NewMailerSpy()
	handler := consumer.NewOfferEmailConsumer(spy)

	event := &domain.BuyerOfferMadeEvent{
		AdID:        "ad-123",
		AdTitle:     "Vélo hollandais",
		SellerEmail: "seller@example.com",
		BuyerEmail:  "buyer@example.com",
		BuyerPseudo: "BuyerPseudo",
		Amount:      15000,
		Message:     "Je suis intéressé, encore disponible ?",
		OccurredAt:  time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC),
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
	if !strings.Contains(email.Body, "BuyerPseudo") {
		t.Errorf("expected body to mention buyer pseudo, got %s", email.Body)
	}
	if !strings.Contains(email.Body, "buyer@example.com") {
		t.Errorf("expected body to mention buyer email, got %s", email.Body)
	}
	if !strings.Contains(email.Body, "150.00") {
		t.Errorf("expected body to mention amount, got %s", email.Body)
	}
	if !strings.Contains(email.Body, "Je suis intéressé, encore disponible ?") {
		t.Errorf("expected body to mention offer message, got %s", email.Body)
	}
}

func TestOfferEmailConsumer_IgnoresOtherEvents(t *testing.T) {
	spy := mailertesting.NewMailerSpy()
	handler := consumer.NewOfferEmailConsumer(spy)

	event := &domain.ClassifiedAdExpiredEvent{AdID: "ad-123"}

	if err := handler(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.GetSentSimpleEmails()) != 0 {
		t.Errorf("expected no email sent for unrelated event")
	}
}

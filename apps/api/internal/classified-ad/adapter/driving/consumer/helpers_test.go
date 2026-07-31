package consumer_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

// fakePasswordHasher is a plaintext-passthrough implementation of
// domain.PasswordHasher for tests: it "hashes" by prefixing the plaintext.
type fakePasswordHasher struct{}

var errFakeHashMismatch = errors.New("hash mismatch")

func (fakePasswordHasher) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

func (fakePasswordHasher) Compare(hash, plain string) error {
	if hash == "hashed:"+plain {
		return nil
	}
	return errFakeHashMismatch
}

// seedSubmittedAd creates and persists a valid classified ad still awaiting
// moderation (StatusSubmitted), owned by seller@example.com, returning it.
func seedSubmittedAd(t *testing.T, repo *inmemory.InMemoryClassifiedAdRepository, submittedAt time.Time) *domain.ClassifiedAd {
	t.Helper()

	email, err := domain.NewEmail("seller@example.com")
	require.NoError(t, err)
	password, err := domain.NewPassword("supersecret", fakePasswordHasher{})
	require.NoError(t, err)
	seller, err := domain.NewSeller(email, "seller-pseudo", password)
	require.NoError(t, err)
	category, err := domain.NewCategory(string(domain.CategoryConsumerGoods))
	require.NoError(t, err)
	location, err := domain.NewLocation("75001", "Paris")
	require.NoError(t, err)

	ad, err := domain.NewClassifiedAd(
		"Vélo hollandais",
		"Très bon état, peu servi.",
		15000,
		seller,
		[]string{"http://img/1.jpg"},
		category,
		location,
		domain.NewSubmissionDate(submittedAt),
	)
	require.NoError(t, err)

	require.NoError(t, repo.Save(ad))
	return ad
}

// mockEvent is a stub eventbus.DomainEvent whose EventType is configurable,
// used to verify that consumers ignore payloads of an unexpected type.
type mockEvent struct {
	eventType string
}

func (e *mockEvent) EventType() string {
	return e.eventType
}

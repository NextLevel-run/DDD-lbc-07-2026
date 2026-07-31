package domain_test

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newValidSnapshot() *domain.ClassifiedAdSnapshot {
	return &domain.ClassifiedAdSnapshot{
		Title:        "A great bike",
		Description:  "Barely used, excellent condition.",
		PriceInCents: 1500,
		ImageURLs:    []string{"http://example.com/img1.jpg"},
		Category:     "consumer_goods",
		ZipCode:      "75001",
		CityName:     "Paris",
		SellerEmail:  "seller@example.com",
		SellerPseudo: "seller-pseudo",
	}
}

func TestNewHistoryAction(t *testing.T) {
	t.Run("valid actions are accepted", func(t *testing.T) {
		for _, valid := range []string{"submitted", "approved", "rejected", "challenged", "published", "deleted", "expired"} {
			action, err := domain.NewHistoryAction(valid)
			assert.NoError(t, err)
			assert.Equal(t, domain.HistoryAction(valid), action)
		}
	})

	t.Run("invalid action is rejected", func(t *testing.T) {
		_, err := domain.NewHistoryAction("archived")
		assert.ErrorIs(t, err, domain.ErrInvalidHistoryAction)
	})

	t.Run("empty action is rejected", func(t *testing.T) {
		_, err := domain.NewHistoryAction("")
		assert.ErrorIs(t, err, domain.ErrInvalidHistoryAction)
	})
}

func TestNewHistoryEntry(t *testing.T) {
	now := time.Now()

	t.Run("submitted entry carries a snapshot", func(t *testing.T) {
		snapshot := newValidSnapshot()

		entry, err := domain.NewHistoryEntry(now, domain.HistoryActionSubmitted, nil, nil, snapshot)

		require.NoError(t, err)
		assert.Equal(t, now, entry.OccurredAt())
		assert.Equal(t, domain.HistoryActionSubmitted, entry.Action())
		assert.Nil(t, entry.ModeratorID())
		assert.Nil(t, entry.Reason())
		assert.Equal(t, snapshot, entry.Snapshot())
	})

	t.Run("moderation entry carries moderator id and reason", func(t *testing.T) {
		moderatorID := "moderator-1"
		reason := "suspect_price"

		entry, err := domain.NewHistoryEntry(now, domain.HistoryActionRejected, &moderatorID, &reason, nil)

		require.NoError(t, err)
		assert.Equal(t, domain.HistoryActionRejected, entry.Action())
		require.NotNil(t, entry.ModeratorID())
		assert.Equal(t, moderatorID, *entry.ModeratorID())
		require.NotNil(t, entry.Reason())
		assert.Equal(t, reason, *entry.Reason())
		assert.Nil(t, entry.Snapshot())
	})

	t.Run("invalid action is rejected", func(t *testing.T) {
		_, err := domain.NewHistoryEntry(now, domain.HistoryAction("archived"), nil, nil, nil)
		assert.ErrorIs(t, err, domain.ErrInvalidHistoryAction)
	})
}

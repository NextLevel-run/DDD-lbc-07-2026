package consumer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
)

func TestClassifiedAdRejectedConsumer_AppendsRejectedEntryWithModeratorAndReason(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdRejectedConsumer(setup.publicBus, setup.appendHistory))
	occurredAt := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

	// When
	err := setup.publicBus.Publish(&shared.ClassifiedAdRejected{
		ClassifiedAdID: "ad-1",
		ModeratorID:    "moderator-1",
		Reason:         "suspect_price",
		OccurredAt:     occurredAt,
	})

	// Then
	require.NoError(t, err)

	history := setup.findHistory(t, "ad-1")
	entries := history.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, domain.HistoryActionRejected, entries[0].Action())
	assert.Equal(t, occurredAt, entries[0].OccurredAt())
	require.NotNil(t, entries[0].ModeratorID())
	assert.Equal(t, "moderator-1", *entries[0].ModeratorID())
	require.NotNil(t, entries[0].Reason())
	assert.Equal(t, "suspect_price", *entries[0].Reason())
	assert.Nil(t, entries[0].Snapshot())
}

func TestClassifiedAdRejectedConsumer_IgnoresUnexpectedEventType(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdRejectedConsumer(setup.publicBus, setup.appendHistory))

	// When
	err := setup.publicBus.Publish(&mockEvent{eventType: shared.ClassifiedAdRejectedEventType})

	// Then
	require.NoError(t, err)
	_, err = setup.historyRepo.FindByClassifiedAdID("ad-1")
	assert.ErrorIs(t, err, domain.ErrClassifiedAdHistoryNotFound)
}

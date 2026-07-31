package consumer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
)

func TestClassifiedAdExpiredConsumer_AppendsExpiredEntry(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdExpiredConsumer(setup.publicBus, setup.appendHistory))
	occurredAt := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)

	// When
	err := setup.publicBus.Publish(&shared.ClassifiedAdExpired{
		ClassifiedAdID: "ad-1",
		OccurredAt:     occurredAt,
	})

	// Then
	require.NoError(t, err)

	history := setup.findHistory(t, "ad-1")
	entries := history.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, domain.HistoryActionExpired, entries[0].Action())
	assert.Equal(t, occurredAt, entries[0].OccurredAt())
	assert.Nil(t, entries[0].ModeratorID())
	assert.Nil(t, entries[0].Reason())
	assert.Nil(t, entries[0].Snapshot())
}

func TestClassifiedAdExpiredConsumer_IgnoresUnexpectedEventType(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdExpiredConsumer(setup.publicBus, setup.appendHistory))

	// When
	err := setup.publicBus.Publish(&mockEvent{eventType: shared.ClassifiedAdExpiredEventType})

	// Then
	require.NoError(t, err)
	_, err = setup.historyRepo.FindByClassifiedAdID("ad-1")
	assert.ErrorIs(t, err, domain.ErrClassifiedAdHistoryNotFound)
}

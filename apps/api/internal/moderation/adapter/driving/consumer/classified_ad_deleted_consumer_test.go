package consumer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
)

func TestClassifiedAdDeletedConsumer_AppendsDeletedEntryWithReason(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdDeletedConsumer(setup.publicBus, setup.appendHistory))
	occurredAt := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

	// When
	err := setup.publicBus.Publish(&shared.ClassifiedAdDeleted{
		ClassifiedAdID: "ad-1",
		Reason:         "rejected",
		OccurredAt:     occurredAt,
	})

	// Then
	require.NoError(t, err)

	history := setup.findHistory(t, "ad-1")
	entries := history.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, domain.HistoryActionDeleted, entries[0].Action())
	assert.Equal(t, occurredAt, entries[0].OccurredAt())
	require.NotNil(t, entries[0].Reason())
	assert.Equal(t, "rejected", *entries[0].Reason())
	assert.Nil(t, entries[0].ModeratorID())
	assert.Nil(t, entries[0].Snapshot())
}

func TestClassifiedAdDeletedConsumer_OmitsEmptyReason(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdDeletedConsumer(setup.publicBus, setup.appendHistory))

	// When
	err := setup.publicBus.Publish(&shared.ClassifiedAdDeleted{
		ClassifiedAdID: "ad-1",
		OccurredAt:     time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC),
	})

	// Then
	require.NoError(t, err)

	history := setup.findHistory(t, "ad-1")
	entries := history.Entries()
	require.Len(t, entries, 1)
	assert.Nil(t, entries[0].Reason())
}

func TestClassifiedAdDeletedConsumer_IgnoresUnexpectedEventType(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdDeletedConsumer(setup.publicBus, setup.appendHistory))

	// When
	err := setup.publicBus.Publish(&mockEvent{eventType: shared.ClassifiedAdDeletedEventType})

	// Then
	require.NoError(t, err)
	_, err = setup.historyRepo.FindByClassifiedAdID("ad-1")
	assert.ErrorIs(t, err, domain.ErrClassifiedAdHistoryNotFound)
}

package consumer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
)

func newEditedEvent(classifiedAdID string) *shared.ClassifiedAdEdited {
	return &shared.ClassifiedAdEdited{
		ClassifiedAdID: classifiedAdID,
		Title:          "Vintage bike (fixed price)",
		Description:    "A sturdy vintage bike, price corrected",
		PriceInCents:   12000,
		ImageURLs:      []string{"https://img.example/1.jpg"},
		Category:       "sports",
		ZipCode:        "75001",
		CityName:       "Paris",
		SellerEmail:    "seller@example.com",
		SellerPseudo:   "seller42",
		OccurredAt:     time.Date(2026, 7, 31, 11, 0, 0, 0, time.UTC),
	}
}

func TestClassifiedAdEditedConsumer_CreatesNewTaskAndSubmittedEntry(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdEditedConsumer(setup.publicBus, setup.createTask, setup.appendHistory))
	event := newEditedEvent("ad-1")

	// When
	err := setup.publicBus.Publish(event)

	// Then - a brand new task is enqueued for the re-submission
	require.NoError(t, err)

	tasks, err := setup.taskRepo.FindAll()
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "ad-1", tasks[0].ClassifiedAdID())

	// The history records a "submitted" entry with the updated snapshot: an
	// edition is a re-submission (there is no "edited" history action).
	history := setup.findHistory(t, "ad-1")
	entries := history.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, domain.HistoryActionSubmitted, entries[0].Action())
	assert.Equal(t, event.OccurredAt, entries[0].OccurredAt())

	snapshot := entries[0].Snapshot()
	require.NotNil(t, snapshot)
	assert.Equal(t, "Vintage bike (fixed price)", snapshot.Title)
	assert.Equal(t, int64(12000), snapshot.PriceInCents)
}

func TestClassifiedAdEditedConsumer_IgnoresUnexpectedEventType(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdEditedConsumer(setup.publicBus, setup.createTask, setup.appendHistory))

	// When
	err := setup.publicBus.Publish(&mockEvent{eventType: shared.ClassifiedAdEditedEventType})

	// Then
	require.NoError(t, err)

	tasks, err := setup.taskRepo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

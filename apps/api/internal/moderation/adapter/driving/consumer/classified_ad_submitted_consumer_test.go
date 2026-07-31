package consumer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
)

func newSubmittedEvent(classifiedAdID string) *shared.ClassifiedAdSubmitted {
	return &shared.ClassifiedAdSubmitted{
		ClassifiedAdID: classifiedAdID,
		Title:          "Vintage bike",
		Description:    "A sturdy vintage bike",
		PriceInCents:   15000,
		ImageURLs:      []string{"https://img.example/1.jpg"},
		Category:       "sports",
		ZipCode:        "75001",
		CityName:       "Paris",
		SellerEmail:    "seller@example.com",
		SellerPseudo:   "seller42",
		OccurredAt:     time.Date(2026, 7, 31, 9, 0, 0, 0, time.UTC),
	}
}

func TestClassifiedAdSubmittedConsumer_CreatesTaskAndHistoryEntry(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdSubmittedConsumer(setup.publicBus, setup.createTask, setup.appendHistory))
	event := newSubmittedEvent("ad-1")

	// When
	err := setup.publicBus.Publish(event)

	// Then
	require.NoError(t, err)

	tasks, err := setup.taskRepo.FindAll()
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "ad-1", tasks[0].ClassifiedAdID())
	assert.False(t, tasks[0].IsClaimed())

	history := setup.findHistory(t, "ad-1")
	entries := history.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, domain.HistoryActionSubmitted, entries[0].Action())
	assert.Equal(t, event.OccurredAt, entries[0].OccurredAt())
	assert.Nil(t, entries[0].ModeratorID())
	assert.Nil(t, entries[0].Reason())

	snapshot := entries[0].Snapshot()
	require.NotNil(t, snapshot)
	assert.Equal(t, "Vintage bike", snapshot.Title)
	assert.Equal(t, "A sturdy vintage bike", snapshot.Description)
	assert.Equal(t, int64(15000), snapshot.PriceInCents)
	assert.Equal(t, []string{"https://img.example/1.jpg"}, snapshot.ImageURLs)
	assert.Equal(t, "sports", snapshot.Category)
	assert.Equal(t, "75001", snapshot.ZipCode)
	assert.Equal(t, "Paris", snapshot.CityName)
	assert.Equal(t, "seller@example.com", snapshot.SellerEmail)
	assert.Equal(t, "seller42", snapshot.SellerPseudo)
}

func TestClassifiedAdSubmittedConsumer_EachSubmissionCreatesANewTask(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdSubmittedConsumer(setup.publicBus, setup.createTask, setup.appendHistory))

	// When - the same ad is submitted twice
	require.NoError(t, setup.publicBus.Publish(newSubmittedEvent("ad-1")))
	require.NoError(t, setup.publicBus.Publish(newSubmittedEvent("ad-1")))

	// Then - two distinct tasks and two history entries exist
	tasks, err := setup.taskRepo.FindAll()
	require.NoError(t, err)
	require.Len(t, tasks, 2)
	assert.NotEqual(t, tasks[0].ID(), tasks[1].ID())

	history := setup.findHistory(t, "ad-1")
	assert.Len(t, history.Entries(), 2)
}

func TestClassifiedAdSubmittedConsumer_IgnoresUnexpectedEventType(t *testing.T) {
	// Given
	setup := newConsumerTestSetup(t)
	require.NoError(t, NewClassifiedAdSubmittedConsumer(setup.publicBus, setup.createTask, setup.appendHistory))

	// When - an event with the right type string but the wrong concrete type
	err := setup.publicBus.Publish(&mockEvent{eventType: shared.ClassifiedAdSubmittedEventType})

	// Then - nothing happens, no failure
	require.NoError(t, err)

	tasks, err := setup.taskRepo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

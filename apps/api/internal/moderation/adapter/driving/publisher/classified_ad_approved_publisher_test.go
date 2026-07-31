package publisher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
)

func TestClassifiedAdApprovedPublisher_RelaysInternalEventToPublicBus(t *testing.T) {
	// Given
	setup := newPublisherTestSetup(t, shared.ClassifiedAdApprovedEventType)
	require.NoError(t, NewClassifiedAdApprovedPublisher(setup.internalBus, setup.publicBus))
	occurredAt := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

	// When
	err := setup.internalBus.Publish(&domain.ClassifiedAdApprovedEvent{
		ClassifiedAdID: "ad-1",
		ModeratorID:    "moderator-1",
		OccurredAt:     occurredAt,
	})

	// Then
	require.NoError(t, err)

	events := setup.collector.GetEvents()
	require.Len(t, events, 1)

	publicEvent, ok := events[0].(*shared.ClassifiedAdApproved)
	require.True(t, ok, "Expected a *shared.ClassifiedAdApproved, got %T", events[0])
	assert.Equal(t, "ad-1", publicEvent.ClassifiedAdID)
	assert.Equal(t, "moderator-1", publicEvent.ModeratorID)
	assert.Equal(t, occurredAt, publicEvent.OccurredAt)
}

func TestClassifiedAdApprovedPublisher_IgnoresUnexpectedEventType(t *testing.T) {
	// Given
	setup := newPublisherTestSetup(t, shared.ClassifiedAdApprovedEventType)
	require.NoError(t, NewClassifiedAdApprovedPublisher(setup.internalBus, setup.publicBus))

	// When - an event with the right type string but the wrong concrete type
	err := setup.internalBus.Publish(&mockEvent{eventType: shared.ClassifiedAdApprovedEventType})

	// Then - nothing is relayed, no failure
	require.NoError(t, err)
	assert.Empty(t, setup.collector.GetEvents())
}

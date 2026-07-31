package publisher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
)

func TestClassifiedAdChallengedPublisher_RelaysInternalEventToPublicBus(t *testing.T) {
	// Given
	setup := newPublisherTestSetup(t, shared.ClassifiedAdChallengedEventType)
	require.NoError(t, NewClassifiedAdChallengedPublisher(setup.internalBus, setup.publicBus))
	occurredAt := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

	// When
	err := setup.internalBus.Publish(&domain.ClassifiedAdChallengedEvent{
		ClassifiedAdID: "ad-1",
		ModeratorID:    "moderator-1",
		Reason:         "price_to_verify",
		OccurredAt:     occurredAt,
	})

	// Then
	require.NoError(t, err)

	events := setup.collector.GetEvents()
	require.Len(t, events, 1)

	publicEvent, ok := events[0].(*shared.ClassifiedAdChallenged)
	require.True(t, ok, "Expected a *shared.ClassifiedAdChallenged, got %T", events[0])
	assert.Equal(t, "ad-1", publicEvent.ClassifiedAdID)
	assert.Equal(t, "moderator-1", publicEvent.ModeratorID)
	assert.Equal(t, "price_to_verify", publicEvent.Reason)
	assert.Equal(t, occurredAt, publicEvent.OccurredAt)
}

func TestClassifiedAdChallengedPublisher_IgnoresUnexpectedEventType(t *testing.T) {
	// Given
	setup := newPublisherTestSetup(t, shared.ClassifiedAdChallengedEventType)
	require.NoError(t, NewClassifiedAdChallengedPublisher(setup.internalBus, setup.publicBus))

	// When
	err := setup.internalBus.Publish(&mockEvent{eventType: shared.ClassifiedAdChallengedEventType})

	// Then
	require.NoError(t, err)
	assert.Empty(t, setup.collector.GetEvents())
}

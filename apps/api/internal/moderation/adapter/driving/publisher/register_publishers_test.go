package publisher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"
)

func TestRegisterPublishers_RelaysAllThreeModerationEvents(t *testing.T) {
	// Given
	internalBus := eventbus.NewSyncInMemoryEventBus()
	publicBus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	for _, eventType := range []string{
		shared.ClassifiedAdApprovedEventType,
		shared.ClassifiedAdRejectedEventType,
		shared.ClassifiedAdChallengedEventType,
	} {
		require.NoError(t, publicBus.Subscribe(eventType, collector.EventHandler()))
	}

	require.NoError(t, RegisterPublishers(internalBus, publicBus))
	occurredAt := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

	// When - each internal decision event is published on the internal bus
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdApprovedEvent{
		ClassifiedAdID: "ad-1", ModeratorID: "moderator-1", OccurredAt: occurredAt,
	}))
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdRejectedEvent{
		ClassifiedAdID: "ad-2", ModeratorID: "moderator-1", Reason: "wrong_category", OccurredAt: occurredAt,
	}))
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdChallengedEvent{
		ClassifiedAdID: "ad-3", ModeratorID: "moderator-1", Reason: "category_to_fix", OccurredAt: occurredAt,
	}))

	// Then - the three public counterparts reached the public bus
	events := collector.GetEvents()
	require.Len(t, events, 3)
	assert.IsType(t, &shared.ClassifiedAdApproved{}, events[0])
	assert.IsType(t, &shared.ClassifiedAdRejected{}, events[1])
	assert.IsType(t, &shared.ClassifiedAdChallenged{}, events[2])
}

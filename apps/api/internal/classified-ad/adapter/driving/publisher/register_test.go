package publisher_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/publisher"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"
)

func TestRegisterPublishers_BridgesAllInternalEventsToPublicBus(t *testing.T) {
	// Given
	internalBus := eventbus.NewSyncInMemoryEventBus()
	publicBus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	for _, eventType := range []string{
		shared.ClassifiedAdSubmittedEventType,
		shared.ClassifiedAdEditedEventType,
		shared.ClassifiedAdPublishedEventType,
		shared.ClassifiedAdDeletedEventType,
		shared.ClassifiedAdExpiredEventType,
	} {
		require.NoError(t, publicBus.Subscribe(eventType, collector.EventHandler()))
	}

	require.NoError(t, publisher.RegisterPublishers(internalBus, publicBus))

	occurredAt := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)

	// When: every bridged internal event is published on the internal bus
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdSubmittedEvent{AdID: "ad-1", OccurredAt: occurredAt}))
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdEditedEvent{AdID: "ad-2", OccurredAt: occurredAt}))
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdPublishedEvent{AdID: "ad-3", PublishedAt: occurredAt}))
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdDeletedEvent{AdID: "ad-4", Reason: "sold", DeletedAt: occurredAt}))
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdExpiredEvent{AdID: "ad-5", ExpiredAt: occurredAt}))

	// Then: each one reached the public bus as its public DTO counterpart
	events := collector.GetEvents()
	require.Len(t, events, 5)
	assert.IsType(t, &shared.ClassifiedAdSubmitted{}, events[0])
	assert.IsType(t, &shared.ClassifiedAdEdited{}, events[1])
	assert.IsType(t, &shared.ClassifiedAdPublished{}, events[2])
	assert.IsType(t, &shared.ClassifiedAdDeleted{}, events[3])
	assert.IsType(t, &shared.ClassifiedAdExpired{}, events[4])
}

func TestRegisterPublishers_DoesNotBridgeModerationDecisionEvents(t *testing.T) {
	// The internal ClassifiedAdApproved/Rejected/Challenged events must NOT be
	// re-published on the public bus by ClassifiedAd: the public events of the
	// same names are owned by the Moderation bounded context.

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

	require.NoError(t, publisher.RegisterPublishers(internalBus, publicBus))

	// When
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdApprovedEvent{AdID: "ad-1"}))
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdRejectedEvent{AdID: "ad-1"}))
	require.NoError(t, internalBus.Publish(&domain.ClassifiedAdChallengedEvent{AdID: "ad-1"}))

	// Then
	assert.Empty(t, collector.GetEvents(), "Expected no moderation decision event to be bridged by ClassifiedAd")
}

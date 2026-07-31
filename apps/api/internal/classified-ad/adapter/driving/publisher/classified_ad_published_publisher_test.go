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

func TestClassifiedAdPublishedPublisher_PublishesPublicEvent(t *testing.T) {
	// Given
	publicBus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, publicBus.Subscribe(shared.ClassifiedAdPublishedEventType, collector.EventHandler()))
	handler := publisher.NewClassifiedAdPublishedPublisher(publicBus)

	publishedAt := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	internalEvent := &domain.ClassifiedAdPublishedEvent{
		AdID:         "ad-123",
		Title:        "Vélo hollandais",
		Category:     "consumer_goods",
		SellerEmail:  "seller@example.com",
		SellerPseudo: "seller-pseudo",
		PublishedAt:  publishedAt,
	}

	// When
	require.NoError(t, handler(internalEvent))

	// Then
	events := collector.GetEvents()
	require.Len(t, events, 1)

	publicEvent, ok := events[0].(*shared.ClassifiedAdPublished)
	require.True(t, ok, "Expected event to be *shared.ClassifiedAdPublished")
	assert.Equal(t, "ad-123", publicEvent.ClassifiedAdID)
	assert.Equal(t, publishedAt, publicEvent.OccurredAt)
}

func TestClassifiedAdPublishedPublisher_IgnoresOtherEvents(t *testing.T) {
	// Given
	publicBus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, publicBus.Subscribe(shared.ClassifiedAdPublishedEventType, collector.EventHandler()))
	handler := publisher.NewClassifiedAdPublishedPublisher(publicBus)

	// When
	require.NoError(t, handler(&domain.ClassifiedAdDeletedEvent{AdID: "ad-123"}))

	// Then
	assert.Empty(t, collector.GetEvents(), "Expected no public event for an unrelated internal event")
}

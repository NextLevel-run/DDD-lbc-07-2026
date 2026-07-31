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

func TestClassifiedAdSubmittedPublisher_PublishesPublicEvent(t *testing.T) {
	// Given
	publicBus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, publicBus.Subscribe(shared.ClassifiedAdSubmittedEventType, collector.EventHandler()))
	handler := publisher.NewClassifiedAdSubmittedPublisher(publicBus)

	occurredAt := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	internalEvent := &domain.ClassifiedAdSubmittedEvent{
		AdID:         "ad-123",
		Title:        "Vélo hollandais",
		Description:  "Très bon état, peu servi.",
		PriceInCents: 15000,
		ImageURLs:    []string{"http://img/1.jpg"},
		Category:     "consumer_goods",
		ZipCode:      "75001",
		CityName:     "Paris",
		SellerEmail:  "seller@example.com",
		SellerPseudo: "seller-pseudo",
		OccurredAt:   occurredAt,
	}

	// When
	require.NoError(t, handler(internalEvent))

	// Then
	events := collector.GetEvents()
	require.Len(t, events, 1)

	publicEvent, ok := events[0].(*shared.ClassifiedAdSubmitted)
	require.True(t, ok, "Expected event to be *shared.ClassifiedAdSubmitted")
	assert.Equal(t, "ad-123", publicEvent.ClassifiedAdID)
	assert.Equal(t, internalEvent.Title, publicEvent.Title)
	assert.Equal(t, internalEvent.Description, publicEvent.Description)
	assert.Equal(t, internalEvent.PriceInCents, publicEvent.PriceInCents)
	assert.Equal(t, internalEvent.ImageURLs, publicEvent.ImageURLs)
	assert.Equal(t, internalEvent.Category, publicEvent.Category)
	assert.Equal(t, internalEvent.ZipCode, publicEvent.ZipCode)
	assert.Equal(t, internalEvent.CityName, publicEvent.CityName)
	assert.Equal(t, internalEvent.SellerEmail, publicEvent.SellerEmail)
	assert.Equal(t, internalEvent.SellerPseudo, publicEvent.SellerPseudo)
	assert.Equal(t, occurredAt, publicEvent.OccurredAt)
}

func TestClassifiedAdSubmittedPublisher_IgnoresOtherEvents(t *testing.T) {
	// Given
	publicBus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, publicBus.Subscribe(shared.ClassifiedAdSubmittedEventType, collector.EventHandler()))
	handler := publisher.NewClassifiedAdSubmittedPublisher(publicBus)

	// When
	require.NoError(t, handler(&domain.ClassifiedAdDeletedEvent{AdID: "ad-123"}))

	// Then
	assert.Empty(t, collector.GetEvents(), "Expected no public event for an unrelated internal event")
}

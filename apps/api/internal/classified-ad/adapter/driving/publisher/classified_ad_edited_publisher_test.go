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

func TestClassifiedAdEditedPublisher_PublishesPublicEvent(t *testing.T) {
	// Given
	publicBus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, publicBus.Subscribe(shared.ClassifiedAdEditedEventType, collector.EventHandler()))
	handler := publisher.NewClassifiedAdEditedPublisher(publicBus)

	occurredAt := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	internalEvent := &domain.ClassifiedAdEditedEvent{
		AdID:         "ad-123",
		Title:        "Vélo hollandais (prix corrigé)",
		Description:  "Prix ajusté suite à la modération.",
		PriceInCents: 12000,
		ImageURLs:    []string{"http://img/1.jpg", "http://img/2.jpg"},
		Category:     "consumer_goods",
		ZipCode:      "69001",
		CityName:     "Lyon",
		SellerEmail:  "seller@example.com",
		SellerPseudo: "seller-pseudo",
		OccurredAt:   occurredAt,
	}

	// When
	require.NoError(t, handler(internalEvent))

	// Then
	events := collector.GetEvents()
	require.Len(t, events, 1)

	publicEvent, ok := events[0].(*shared.ClassifiedAdEdited)
	require.True(t, ok, "Expected event to be *shared.ClassifiedAdEdited")
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

func TestClassifiedAdEditedPublisher_IgnoresOtherEvents(t *testing.T) {
	// Given
	publicBus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, publicBus.Subscribe(shared.ClassifiedAdEditedEventType, collector.EventHandler()))
	handler := publisher.NewClassifiedAdEditedPublisher(publicBus)

	// When
	require.NoError(t, handler(&domain.ClassifiedAdSubmittedEvent{AdID: "ad-123"}))

	// Then
	assert.Empty(t, collector.GetEvents(), "Expected no public event for an unrelated internal event")
}

package publisher

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// RegisterPublishers subscribes every ClassifiedAd publisher on the internal
// event bus, so that the corresponding public integration events reach the
// public bus. It is meant to be called once at wiring time.
func RegisterPublishers(internalBus eventbus.Bus, publicBus shared.PublicEventBus) error {
	subscriptions := []struct {
		eventType string
		handler   eventbus.EventHandler
	}{
		{(&domain.ClassifiedAdSubmittedEvent{}).EventType(), NewClassifiedAdSubmittedPublisher(publicBus)},
		{(&domain.ClassifiedAdEditedEvent{}).EventType(), NewClassifiedAdEditedPublisher(publicBus)},
		{(&domain.ClassifiedAdPublishedEvent{}).EventType(), NewClassifiedAdPublishedPublisher(publicBus)},
		{(&domain.ClassifiedAdDeletedEvent{}).EventType(), NewClassifiedAdDeletedPublisher(publicBus)},
		{(&domain.ClassifiedAdExpiredEvent{}).EventType(), NewClassifiedAdExpiredPublisher(publicBus)},
	}

	for _, subscription := range subscriptions {
		if err := internalBus.Subscribe(subscription.eventType, subscription.handler); err != nil {
			return err
		}
	}

	return nil
}

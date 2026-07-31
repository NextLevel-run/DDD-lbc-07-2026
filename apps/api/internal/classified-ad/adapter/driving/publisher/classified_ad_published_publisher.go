package publisher

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdPublishedPublisher builds an event handler that maps the
// internal ClassifiedAdPublishedEvent to the public ClassifiedAdPublished
// integration event and publishes it on the public bus.
func NewClassifiedAdPublishedPublisher(publicBus shared.PublicEventBus) eventbus.EventHandler {
	return func(event eventbus.DomainEvent) error {
		publishedEvent, ok := event.(*domain.ClassifiedAdPublishedEvent)
		if !ok {
			return nil
		}

		return publicBus.Publish(&shared.ClassifiedAdPublished{
			ClassifiedAdID: publishedEvent.AdID,
			OccurredAt:     publishedEvent.PublishedAt,
		})
	}
}

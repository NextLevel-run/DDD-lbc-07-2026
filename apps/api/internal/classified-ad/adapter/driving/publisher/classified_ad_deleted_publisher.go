package publisher

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdDeletedPublisher builds an event handler that maps the
// internal ClassifiedAdDeletedEvent to the public ClassifiedAdDeleted
// integration event and publishes it on the public bus.
func NewClassifiedAdDeletedPublisher(publicBus shared.PublicEventBus) eventbus.EventHandler {
	return func(event eventbus.DomainEvent) error {
		deletedEvent, ok := event.(*domain.ClassifiedAdDeletedEvent)
		if !ok {
			return nil
		}

		return publicBus.Publish(&shared.ClassifiedAdDeleted{
			ClassifiedAdID: deletedEvent.AdID,
			Reason:         deletedEvent.Reason,
			OccurredAt:     deletedEvent.DeletedAt,
		})
	}
}

package publisher

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdExpiredPublisher builds an event handler that maps the
// internal ClassifiedAdExpiredEvent to the public ClassifiedAdExpired
// integration event and publishes it on the public bus.
func NewClassifiedAdExpiredPublisher(publicBus shared.PublicEventBus) eventbus.EventHandler {
	return func(event eventbus.DomainEvent) error {
		expiredEvent, ok := event.(*domain.ClassifiedAdExpiredEvent)
		if !ok {
			return nil
		}

		return publicBus.Publish(&shared.ClassifiedAdExpired{
			ClassifiedAdID: expiredEvent.AdID,
			OccurredAt:     expiredEvent.ExpiredAt,
		})
	}
}

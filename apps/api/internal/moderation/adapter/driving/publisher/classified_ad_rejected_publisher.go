package publisher

import (
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdRejectedPublisher subscribes to Moderation's internal
// ClassifiedAdRejected event and relays it as the public ClassifiedAdRejected
// integration event on the public bus.
func NewClassifiedAdRejectedPublisher(internalBus eventbus.Bus, publicBus shared.PublicEventBus) error {
	return internalBus.Subscribe(shared.ClassifiedAdRejectedEventType, func(event eventbus.DomainEvent) error {
		rejectedEvent, ok := event.(*domain.ClassifiedAdRejectedEvent)
		if !ok {
			return nil
		}

		return publicBus.Publish(&shared.ClassifiedAdRejected{
			ClassifiedAdID: rejectedEvent.ClassifiedAdID,
			ModeratorID:    rejectedEvent.ModeratorID,
			Reason:         rejectedEvent.Reason,
			OccurredAt:     rejectedEvent.OccurredAt,
		})
	})
}

package publisher

import (
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdApprovedPublisher subscribes to Moderation's internal
// ClassifiedAdApproved event and relays it as the public ClassifiedAdApproved
// integration event on the public bus.
func NewClassifiedAdApprovedPublisher(internalBus eventbus.Bus, publicBus shared.PublicEventBus) error {
	return internalBus.Subscribe(shared.ClassifiedAdApprovedEventType, func(event eventbus.DomainEvent) error {
		approvedEvent, ok := event.(*domain.ClassifiedAdApprovedEvent)
		if !ok {
			return nil
		}

		return publicBus.Publish(&shared.ClassifiedAdApproved{
			ClassifiedAdID: approvedEvent.ClassifiedAdID,
			ModeratorID:    approvedEvent.ModeratorID,
			OccurredAt:     approvedEvent.OccurredAt,
		})
	})
}

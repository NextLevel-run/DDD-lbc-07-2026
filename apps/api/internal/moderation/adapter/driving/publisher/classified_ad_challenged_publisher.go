package publisher

import (
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdChallengedPublisher subscribes to Moderation's internal
// ClassifiedAdChallenged event and relays it as the public
// ClassifiedAdChallenged integration event on the public bus.
func NewClassifiedAdChallengedPublisher(internalBus eventbus.Bus, publicBus shared.PublicEventBus) error {
	return internalBus.Subscribe(shared.ClassifiedAdChallengedEventType, func(event eventbus.DomainEvent) error {
		challengedEvent, ok := event.(*domain.ClassifiedAdChallengedEvent)
		if !ok {
			return nil
		}

		return publicBus.Publish(&shared.ClassifiedAdChallenged{
			ClassifiedAdID: challengedEvent.ClassifiedAdID,
			ModeratorID:    challengedEvent.ModeratorID,
			Reason:         challengedEvent.Reason,
			OccurredAt:     challengedEvent.OccurredAt,
		})
	})
}

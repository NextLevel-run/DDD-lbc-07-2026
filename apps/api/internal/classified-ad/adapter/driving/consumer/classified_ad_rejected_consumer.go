package consumer

import (
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdRejectedConsumer subscribes to the public ClassifiedAdRejected
// event (emitted by the Moderation bounded context) and applies the decision by
// calling RejectClassifiedAdCommand — a system command that transitions the ad
// submitted → rejected and then automatically deletes it (reason "rejected").
func NewClassifiedAdRejectedConsumer(
	publicBus shared.PublicEventBus,
	rejectClassifiedAd command.RejectClassifiedAdCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdRejectedEventType, func(event eventbus.DomainEvent) error {
		rejectedEvent, ok := event.(*shared.ClassifiedAdRejected)
		if !ok {
			return nil
		}

		return rejectClassifiedAd(command.RejectClassifiedAdCommandArgs{
			AdID: rejectedEvent.ClassifiedAdID,
		})
	})
}

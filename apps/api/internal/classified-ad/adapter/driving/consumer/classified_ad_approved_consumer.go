package consumer

import (
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdApprovedConsumer subscribes to the public ClassifiedAdApproved
// event (emitted by the Moderation bounded context) and applies the decision by
// calling ApproveClassifiedAdCommand (submitted → approved).
func NewClassifiedAdApprovedConsumer(
	publicBus shared.PublicEventBus,
	approveClassifiedAd command.ApproveClassifiedAdCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdApprovedEventType, func(event eventbus.DomainEvent) error {
		approvedEvent, ok := event.(*shared.ClassifiedAdApproved)
		if !ok {
			return nil
		}

		return approveClassifiedAd(command.ApproveClassifiedAdCommandArgs{
			AdID: approvedEvent.ClassifiedAdID,
		})
	})
}

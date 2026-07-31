package consumer

import (
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdApprovedInternalConsumer subscribes to the INTERNAL
// ClassifiedAdApproved event and immediately chains the approved → published
// transition by calling PublishClassifiedAdCommand, so an approved ad goes
// online without any extra step.
func NewClassifiedAdApprovedInternalConsumer(
	internalBus eventbus.Bus,
	publishClassifiedAd command.PublishClassifiedAdCommand,
) error {
	return internalBus.Subscribe((&domain.ClassifiedAdApprovedEvent{}).EventType(), func(event eventbus.DomainEvent) error {
		approvedEvent, ok := event.(*domain.ClassifiedAdApprovedEvent)
		if !ok {
			return nil
		}

		return publishClassifiedAd(command.PublishClassifiedAdCommandArgs{
			AdID: approvedEvent.AdID,
		})
	})
}

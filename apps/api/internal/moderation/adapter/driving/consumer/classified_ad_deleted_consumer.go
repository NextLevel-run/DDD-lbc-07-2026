package consumer

import (
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdDeletedConsumer subscribes to the public ClassifiedAdDeleted
// event and appends a "deleted" entry to the ad's history, recording the
// deletion reason carried by the event.
func NewClassifiedAdDeletedConsumer(
	publicBus eventbus.Bus,
	appendHistoryEntry command.AppendHistoryEntryCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdDeletedEventType, func(event eventbus.DomainEvent) error {
		deletedEvent, ok := event.(*shared.ClassifiedAdDeleted)
		if !ok {
			return nil
		}

		var reason *string
		if deletedEvent.Reason != "" {
			reason = &deletedEvent.Reason
		}

		return appendHistoryEntry(command.AppendHistoryEntryCommandArgs{
			ClassifiedAdID: deletedEvent.ClassifiedAdID,
			OccurredAt:     deletedEvent.OccurredAt,
			Action:         string(domain.HistoryActionDeleted),
			Reason:         reason,
		})
	})
}

package consumer

import (
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdRejectedConsumer subscribes to the public ClassifiedAdRejected
// event — Moderation's own outbound event, consumed back for the audit trail —
// and appends a "rejected" entry (with the moderator ID and reason) to the
// ad's history.
func NewClassifiedAdRejectedConsumer(
	publicBus eventbus.Bus,
	appendHistoryEntry command.AppendHistoryEntryCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdRejectedEventType, func(event eventbus.DomainEvent) error {
		rejectedEvent, ok := event.(*shared.ClassifiedAdRejected)
		if !ok {
			return nil
		}

		return appendHistoryEntry(command.AppendHistoryEntryCommandArgs{
			ClassifiedAdID: rejectedEvent.ClassifiedAdID,
			OccurredAt:     rejectedEvent.OccurredAt,
			Action:         string(domain.HistoryActionRejected),
			ModeratorID:    &rejectedEvent.ModeratorID,
			Reason:         &rejectedEvent.Reason,
		})
	})
}

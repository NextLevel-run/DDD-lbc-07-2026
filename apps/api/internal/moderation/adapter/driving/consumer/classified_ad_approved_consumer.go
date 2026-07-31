package consumer

import (
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdApprovedConsumer subscribes to the public ClassifiedAdApproved
// event — Moderation's own outbound event, consumed back for the audit trail —
// and appends an "approved" entry (with the moderator ID) to the ad's history.
func NewClassifiedAdApprovedConsumer(
	publicBus eventbus.Bus,
	appendHistoryEntry command.AppendHistoryEntryCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdApprovedEventType, func(event eventbus.DomainEvent) error {
		approvedEvent, ok := event.(*shared.ClassifiedAdApproved)
		if !ok {
			return nil
		}

		return appendHistoryEntry(command.AppendHistoryEntryCommandArgs{
			ClassifiedAdID: approvedEvent.ClassifiedAdID,
			OccurredAt:     approvedEvent.OccurredAt,
			Action:         string(domain.HistoryActionApproved),
			ModeratorID:    &approvedEvent.ModeratorID,
		})
	})
}

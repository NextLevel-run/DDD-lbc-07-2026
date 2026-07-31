package consumer

import (
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdChallengedConsumer subscribes to the public
// ClassifiedAdChallenged event — Moderation's own outbound event, consumed back
// for the audit trail — and appends a "challenged" entry (with the moderator ID
// and reason) to the ad's history.
func NewClassifiedAdChallengedConsumer(
	publicBus eventbus.Bus,
	appendHistoryEntry command.AppendHistoryEntryCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdChallengedEventType, func(event eventbus.DomainEvent) error {
		challengedEvent, ok := event.(*shared.ClassifiedAdChallenged)
		if !ok {
			return nil
		}

		return appendHistoryEntry(command.AppendHistoryEntryCommandArgs{
			ClassifiedAdID: challengedEvent.ClassifiedAdID,
			OccurredAt:     challengedEvent.OccurredAt,
			Action:         string(domain.HistoryActionChallenged),
			ModeratorID:    &challengedEvent.ModeratorID,
			Reason:         &challengedEvent.Reason,
		})
	})
}

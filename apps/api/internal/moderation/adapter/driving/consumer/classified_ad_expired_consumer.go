package consumer

import (
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdExpiredConsumer subscribes to the public ClassifiedAdExpired
// event and appends an "expired" entry to the ad's history.
func NewClassifiedAdExpiredConsumer(
	publicBus eventbus.Bus,
	appendHistoryEntry command.AppendHistoryEntryCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdExpiredEventType, func(event eventbus.DomainEvent) error {
		expiredEvent, ok := event.(*shared.ClassifiedAdExpired)
		if !ok {
			return nil
		}

		return appendHistoryEntry(command.AppendHistoryEntryCommandArgs{
			ClassifiedAdID: expiredEvent.ClassifiedAdID,
			OccurredAt:     expiredEvent.OccurredAt,
			Action:         string(domain.HistoryActionExpired),
		})
	})
}

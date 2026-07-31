package consumer

import (
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdPublishedConsumer subscribes to the public ClassifiedAdPublished
// event and appends a "published" entry to the ad's history.
func NewClassifiedAdPublishedConsumer(
	publicBus eventbus.Bus,
	appendHistoryEntry command.AppendHistoryEntryCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdPublishedEventType, func(event eventbus.DomainEvent) error {
		publishedEvent, ok := event.(*shared.ClassifiedAdPublished)
		if !ok {
			return nil
		}

		return appendHistoryEntry(command.AppendHistoryEntryCommandArgs{
			ClassifiedAdID: publishedEvent.ClassifiedAdID,
			OccurredAt:     publishedEvent.OccurredAt,
			Action:         string(domain.HistoryActionPublished),
		})
	})
}

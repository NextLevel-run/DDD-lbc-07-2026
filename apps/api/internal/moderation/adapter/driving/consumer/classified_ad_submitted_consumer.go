package consumer

import (
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdSubmittedConsumer subscribes to the public ClassifiedAdSubmitted
// event: each submission enqueues a brand new moderation task and appends a
// "submitted" entry (with the full ad snapshot) to the ad's history.
func NewClassifiedAdSubmittedConsumer(
	publicBus eventbus.Bus,
	createModerationTask command.CreateModerationTaskCommand,
	appendHistoryEntry command.AppendHistoryEntryCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdSubmittedEventType, func(event eventbus.DomainEvent) error {
		submittedEvent, ok := event.(*shared.ClassifiedAdSubmitted)
		if !ok {
			return nil
		}

		if _, err := createModerationTask(command.CreateModerationTaskCommandArgs{
			ClassifiedAdID: submittedEvent.ClassifiedAdID,
		}); err != nil {
			return err
		}

		return appendHistoryEntry(command.AppendHistoryEntryCommandArgs{
			ClassifiedAdID: submittedEvent.ClassifiedAdID,
			OccurredAt:     submittedEvent.OccurredAt,
			Action:         string(domain.HistoryActionSubmitted),
			Snapshot: &domain.ClassifiedAdSnapshot{
				Title:        submittedEvent.Title,
				Description:  submittedEvent.Description,
				PriceInCents: submittedEvent.PriceInCents,
				ImageURLs:    submittedEvent.ImageURLs,
				Category:     submittedEvent.Category,
				ZipCode:      submittedEvent.ZipCode,
				CityName:     submittedEvent.CityName,
				SellerEmail:  submittedEvent.SellerEmail,
				SellerPseudo: submittedEvent.SellerPseudo,
			},
		})
	})
}

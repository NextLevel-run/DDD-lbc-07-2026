package consumer

import (
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdEditedConsumer subscribes to the public ClassifiedAdEdited
// event: each re-submission after a challenge enqueues a brand new moderation
// task (new ID) and appends a "submitted" entry (with the updated ad snapshot)
// to the ad's history — the history action enum has no "edited" value, an
// edition is a re-submission.
func NewClassifiedAdEditedConsumer(
	publicBus eventbus.Bus,
	createModerationTask command.CreateModerationTaskCommand,
	appendHistoryEntry command.AppendHistoryEntryCommand,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdEditedEventType, func(event eventbus.DomainEvent) error {
		editedEvent, ok := event.(*shared.ClassifiedAdEdited)
		if !ok {
			return nil
		}

		if _, err := createModerationTask(command.CreateModerationTaskCommandArgs{
			ClassifiedAdID: editedEvent.ClassifiedAdID,
		}); err != nil {
			return err
		}

		return appendHistoryEntry(command.AppendHistoryEntryCommandArgs{
			ClassifiedAdID: editedEvent.ClassifiedAdID,
			OccurredAt:     editedEvent.OccurredAt,
			Action:         string(domain.HistoryActionSubmitted),
			Snapshot: &domain.ClassifiedAdSnapshot{
				Title:        editedEvent.Title,
				Description:  editedEvent.Description,
				PriceInCents: editedEvent.PriceInCents,
				ImageURLs:    editedEvent.ImageURLs,
				Category:     editedEvent.Category,
				ZipCode:      editedEvent.ZipCode,
				CityName:     editedEvent.CityName,
				SellerEmail:  editedEvent.SellerEmail,
				SellerPseudo: editedEvent.SellerPseudo,
			},
		})
	})
}

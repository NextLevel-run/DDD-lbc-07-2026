package publisher

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdEditedPublisher builds an event handler that maps the
// internal ClassifiedAdEditedEvent to the public ClassifiedAdEdited
// integration event and publishes it on the public bus.
func NewClassifiedAdEditedPublisher(publicBus shared.PublicEventBus) eventbus.EventHandler {
	return func(event eventbus.DomainEvent) error {
		editedEvent, ok := event.(*domain.ClassifiedAdEditedEvent)
		if !ok {
			return nil
		}

		return publicBus.Publish(&shared.ClassifiedAdEdited{
			ClassifiedAdID: editedEvent.AdID,
			Title:          editedEvent.Title,
			Description:    editedEvent.Description,
			PriceInCents:   editedEvent.PriceInCents,
			ImageURLs:      editedEvent.ImageURLs,
			Category:       editedEvent.Category,
			ZipCode:        editedEvent.ZipCode,
			CityName:       editedEvent.CityName,
			SellerEmail:    editedEvent.SellerEmail,
			SellerPseudo:   editedEvent.SellerPseudo,
			OccurredAt:     editedEvent.OccurredAt,
		})
	}
}

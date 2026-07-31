// Package publisher contains the driving adapters that bridge the
// ClassifiedAd internal event bus to the public integration bus: each
// publisher consumes one internal domain event and publishes the
// corresponding public event DTO (defined in internal/shared).
package publisher

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// NewClassifiedAdSubmittedPublisher builds an event handler that maps the
// internal ClassifiedAdSubmittedEvent to the public ClassifiedAdSubmitted
// integration event and publishes it on the public bus.
func NewClassifiedAdSubmittedPublisher(publicBus shared.PublicEventBus) eventbus.EventHandler {
	return func(event eventbus.DomainEvent) error {
		submittedEvent, ok := event.(*domain.ClassifiedAdSubmittedEvent)
		if !ok {
			return nil
		}

		return publicBus.Publish(&shared.ClassifiedAdSubmitted{
			ClassifiedAdID: submittedEvent.AdID,
			Title:          submittedEvent.Title,
			Description:    submittedEvent.Description,
			PriceInCents:   submittedEvent.PriceInCents,
			ImageURLs:      submittedEvent.ImageURLs,
			Category:       submittedEvent.Category,
			ZipCode:        submittedEvent.ZipCode,
			CityName:       submittedEvent.CityName,
			SellerEmail:    submittedEvent.SellerEmail,
			SellerPseudo:   submittedEvent.SellerPseudo,
			OccurredAt:     submittedEvent.OccurredAt,
		})
	}
}

package consumer

import (
	"fmt"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	"ddd-second-hand-marketplace/pkg/mailer"
)

// NewOfferEmailConsumer builds an event handler that emails the seller
// when a buyer makes an offer on their classified ad.
func NewOfferEmailConsumer(m mailer.Mailer) eventbus.EventHandler {
	return func(event eventbus.DomainEvent) error {
		offerEvent, ok := event.(*domain.BuyerOfferMadeEvent)
		if !ok {
			return nil
		}

		subject := fmt.Sprintf("Nouvelle offre sur votre annonce %q", offerEvent.AdTitle)
		body := fmt.Sprintf(
			"Bonjour,\n\n%s (%s) vous a fait une offre de %.2f € sur votre annonce %q :\n\n%s\n\nL'équipe Marketplace",
			offerEvent.BuyerPseudo,
			offerEvent.BuyerEmail,
			float64(offerEvent.Amount)/100,
			offerEvent.AdTitle,
			offerEvent.Message,
		)

		return m.SendSimpleEmail(offerEvent.SellerEmail, fromEmail, subject, body)
	}
}

package consumer

import (
	"fmt"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	"ddd-second-hand-marketplace/pkg/mailer"
)

const fromEmail = "no-reply@marketplace.local"

// NewAdPublishedEmailConsumer builds an event handler that emails the seller
// a confirmation when their classified ad is published.
func NewAdPublishedEmailConsumer(m mailer.Mailer) eventbus.EventHandler {
	return func(event eventbus.DomainEvent) error {
		publishedEvent, ok := event.(*domain.ClassifiedAdPublishedEvent)
		if !ok {
			return nil
		}

		subject := fmt.Sprintf("Votre annonce %q est publiée", publishedEvent.Title)
		body := fmt.Sprintf(
			"Bonjour %s,\n\nVotre annonce %q (catégorie %s) a bien été publiée le %s.\n\nL'équipe Marketplace",
			publishedEvent.SellerPseudo,
			publishedEvent.Title,
			publishedEvent.Category,
			publishedEvent.PublishedAt.Format("02/01/2006 15:04"),
		)

		return m.SendSimpleEmail(publishedEvent.SellerEmail, fromEmail, subject, body)
	}
}

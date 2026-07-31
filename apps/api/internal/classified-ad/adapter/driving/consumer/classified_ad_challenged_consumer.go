package consumer

import (
	"fmt"

	"github.com/google/uuid"

	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
	"ddd-second-hand-marketplace/pkg/mailer"
)

// NewClassifiedAdChallengedConsumer subscribes to the public
// ClassifiedAdChallenged event (emitted by the Moderation bounded context),
// applies the decision by calling ChallengeClassifiedAdCommand (submitted →
// challenged) and emails the seller so they can correct and re-submit their
// ad. The seller's email is looked up through the repository, since the
// public event does not carry it.
func NewClassifiedAdChallengedConsumer(
	publicBus shared.PublicEventBus,
	challengeClassifiedAd command.ChallengeClassifiedAdCommand,
	repo domain.ClassifiedAdRepository,
	m mailer.Mailer,
) error {
	return publicBus.Subscribe(shared.ClassifiedAdChallengedEventType, func(event eventbus.DomainEvent) error {
		challengedEvent, ok := event.(*shared.ClassifiedAdChallenged)
		if !ok {
			return nil
		}

		if err := challengeClassifiedAd(command.ChallengeClassifiedAdCommandArgs{
			AdID: challengedEvent.ClassifiedAdID,
		}); err != nil {
			return err
		}

		// The command succeeded, so the id is well-formed and the ad exists.
		adID, err := uuid.Parse(challengedEvent.ClassifiedAdID)
		if err != nil {
			return err
		}
		ad, err := repo.FindByID(adID)
		if err != nil {
			return err
		}

		subject := fmt.Sprintf("Votre annonce %q nécessite des corrections", ad.Title())
		body := fmt.Sprintf(
			"Bonjour %s,\n\nVotre annonce %q a été examinée par la modération et nécessite des corrections (motif : %s).\n\nMerci de modifier votre annonce pour la soumettre à nouveau.\n\nL'équipe Marketplace",
			ad.Seller().Pseudo(),
			ad.Title(),
			challengedEvent.Reason,
		)

		return m.SendSimpleEmail(ad.Seller().Email().String(), fromEmail, subject, body)
	})
}

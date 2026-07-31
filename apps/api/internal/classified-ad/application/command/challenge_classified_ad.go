package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// ChallengeClassifiedAdCommandArgs contains input data for the command.
type ChallengeClassifiedAdCommandArgs struct {
	AdID string
}

// ChallengeClassifiedAdCommand is the command function type. It is a system
// command triggered by the moderation challenge decision — no seller
// credentials are involved. The seller must then edit the ad to re-submit it.
type ChallengeClassifiedAdCommand func(args ChallengeClassifiedAdCommandArgs) error

// BuildChallengeClassifiedAdCommand builds a command with dependencies injected.
func BuildChallengeClassifiedAdCommand(
	repo domain.ClassifiedAdRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) ChallengeClassifiedAdCommand {
	return func(args ChallengeClassifiedAdCommandArgs) error {
		adID, err := uuid.Parse(args.AdID)
		if err != nil {
			return domain.ErrClassifiedAdNotFound
		}

		ad, err := repo.FindByID(adID)
		if err != nil {
			return err
		}

		if err := ad.Challenge(); err != nil {
			return err
		}

		if err := repo.Save(ad); err != nil {
			return err
		}

		event := domain.NewClassifiedAdChallengedEventFromClassifiedAd(ad, clock.Now())
		return eventBus.Publish(event)
	}
}

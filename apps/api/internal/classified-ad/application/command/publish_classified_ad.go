package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// PublishClassifiedAdCommandArgs contains input data for the command.
type PublishClassifiedAdCommandArgs struct {
	AdID string
}

// PublishClassifiedAdCommand is the command function type. It is a system
// command chained immediately after moderation approval: it puts an approved
// ad online, setting its publication date.
type PublishClassifiedAdCommand func(args PublishClassifiedAdCommandArgs) error

// BuildPublishClassifiedAdCommand builds a command with dependencies injected.
func BuildPublishClassifiedAdCommand(
	repo domain.ClassifiedAdRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) PublishClassifiedAdCommand {
	return func(args PublishClassifiedAdCommandArgs) error {
		adID, err := uuid.Parse(args.AdID)
		if err != nil {
			return domain.ErrClassifiedAdNotFound
		}

		ad, err := repo.FindByID(adID)
		if err != nil {
			return err
		}

		if err := ad.Publish(clock.Now()); err != nil {
			return err
		}

		if err := repo.Save(ad); err != nil {
			return err
		}

		event := domain.NewClassifiedAdPublishedEventFromClassifiedAd(ad)
		return eventBus.Publish(event)
	}
}

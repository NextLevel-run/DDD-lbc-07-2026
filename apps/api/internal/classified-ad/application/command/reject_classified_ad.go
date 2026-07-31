package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// RejectClassifiedAdCommandArgs contains input data for the command.
type RejectClassifiedAdCommandArgs struct {
	AdID string
}

// RejectClassifiedAdCommand is the command function type. It is a system
// command triggered by the moderation rejection decision — no seller
// credentials are involved. It applies the submitted → rejected transition
// followed by the automatic rejected → deleted transition (delete reason
// "rejected").
type RejectClassifiedAdCommand func(args RejectClassifiedAdCommandArgs) error

// BuildRejectClassifiedAdCommand builds a command with dependencies injected.
func BuildRejectClassifiedAdCommand(
	repo domain.ClassifiedAdRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) RejectClassifiedAdCommand {
	return func(args RejectClassifiedAdCommandArgs) error {
		adID, err := uuid.Parse(args.AdID)
		if err != nil {
			return domain.ErrClassifiedAdNotFound
		}

		ad, err := repo.FindByID(adID)
		if err != nil {
			return err
		}

		now := clock.Now()

		if err := ad.Reject(); err != nil {
			return err
		}

		if err := ad.DeleteRejected(now); err != nil {
			return err
		}

		if err := repo.Save(ad); err != nil {
			return err
		}

		rejectedEvent := domain.NewClassifiedAdRejectedEventFromClassifiedAd(ad, now)
		if err := eventBus.Publish(rejectedEvent); err != nil {
			return err
		}

		deletedEvent := domain.NewClassifiedAdDeletedEventFromClassifiedAd(ad)
		return eventBus.Publish(deletedEvent)
	}
}

package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// ApproveClassifiedAdCommandArgs contains input data for the command.
type ApproveClassifiedAdCommandArgs struct {
	AdID string
}

// ApproveClassifiedAdCommand is the command function type. It is a system
// command triggered by the moderation approval decision — no seller
// credentials are involved.
type ApproveClassifiedAdCommand func(args ApproveClassifiedAdCommandArgs) error

// BuildApproveClassifiedAdCommand builds a command with dependencies injected.
func BuildApproveClassifiedAdCommand(
	repo domain.ClassifiedAdRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) ApproveClassifiedAdCommand {
	return func(args ApproveClassifiedAdCommandArgs) error {
		adID, err := uuid.Parse(args.AdID)
		if err != nil {
			return domain.ErrClassifiedAdNotFound
		}

		ad, err := repo.FindByID(adID)
		if err != nil {
			return err
		}

		if err := ad.Approve(); err != nil {
			return err
		}

		if err := repo.Save(ad); err != nil {
			return err
		}

		event := domain.NewClassifiedAdApprovedEventFromClassifiedAd(ad, clock.Now())
		return eventBus.Publish(event)
	}
}

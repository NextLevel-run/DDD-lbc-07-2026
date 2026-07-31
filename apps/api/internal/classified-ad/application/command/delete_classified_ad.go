package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// DeleteClassifiedAdCommandArgs contains input data for the command.
type DeleteClassifiedAdCommandArgs struct {
	AdID     string
	Email    string
	Password string
	Reason   string
}

// DeleteClassifiedAdCommand is the command function type.
type DeleteClassifiedAdCommand func(args DeleteClassifiedAdCommandArgs) error

// BuildDeleteClassifiedAdCommand builds a command with dependencies injected.
func BuildDeleteClassifiedAdCommand(
	repo domain.ClassifiedAdRepository,
	hasher domain.PasswordHasher,
	clock domain.Clock,
	eventBus eventbus.Bus,
) DeleteClassifiedAdCommand {
	return func(args DeleteClassifiedAdCommandArgs) error {
		adID, err := uuid.Parse(args.AdID)
		if err != nil {
			return domain.ErrClassifiedAdNotFound
		}

		ad, err := repo.FindByID(adID)
		if err != nil {
			return err
		}

		email, err := domain.NewEmail(args.Email)
		if err != nil {
			return err
		}

		reason, err := domain.NewDeleteReason(args.Reason)
		if err != nil {
			return err
		}

		deleted, err := ad.Delete(email, args.Password, reason, hasher, clock.Now())
		if err != nil {
			return err
		}

		if !deleted {
			return nil
		}

		if err := repo.Save(ad); err != nil {
			return err
		}

		event := domain.NewClassifiedAdDeletedEventFromClassifiedAd(ad)
		return eventBus.Publish(event)
	}
}

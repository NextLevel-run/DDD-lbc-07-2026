package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// EditClassifiedAdCommandArgs contains input data for the command.
type EditClassifiedAdCommandArgs struct {
	AdID         string
	Email        string
	Password     string
	Title        string
	Description  string
	PriceInCents int64
	ImageURLs    []string
	Category     string
	ZipCode      string
	CityName     string
}

// EditClassifiedAdCommand is the command function type. The seller edits a
// challenged ad (all fields are editable), which re-submits it for moderation.
type EditClassifiedAdCommand func(args EditClassifiedAdCommandArgs) error

// BuildEditClassifiedAdCommand builds a command with dependencies injected.
func BuildEditClassifiedAdCommand(
	repo domain.ClassifiedAdRepository,
	hasher domain.PasswordHasher,
	clock domain.Clock,
	eventBus eventbus.Bus,
) EditClassifiedAdCommand {
	return func(args EditClassifiedAdCommandArgs) error {
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

		// Same authentication mechanism as DeleteClassifiedAdCommand: the
		// seller must provide their email and password.
		if email.String() != ad.Seller().Email().String() || !ad.Seller().Password().Matches(args.Password, hasher) {
			return domain.ErrInvalidCredentials
		}

		category, err := domain.NewCategory(args.Category)
		if err != nil {
			return err
		}

		location, err := domain.NewLocation(args.ZipCode, args.CityName)
		if err != nil {
			return err
		}

		if err := ad.Edit(args.Title, args.Description, args.PriceInCents, args.ImageURLs, category, location); err != nil {
			return err
		}

		if err := repo.Save(ad); err != nil {
			return err
		}

		event := domain.NewClassifiedAdEditedEventFromClassifiedAd(ad, clock.Now())
		return eventBus.Publish(event)
	}
}

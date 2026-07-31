package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// SubmitClassifiedAdCommandArgs contains input data for the command.
type SubmitClassifiedAdCommandArgs struct {
	Title          string
	Description    string
	PriceInCents   int64
	SellerEmail    string
	SellerPseudo   string
	SellerPassword string
	ImageURLs      []string
	Category       string
	ZipCode        string
	CityName       string
}

// SubmitClassifiedAdCommand is the command function type.
type SubmitClassifiedAdCommand func(args SubmitClassifiedAdCommandArgs) (string, error)

// BuildSubmitClassifiedAdCommand builds a command with dependencies injected.
func BuildSubmitClassifiedAdCommand(
	repo domain.ClassifiedAdRepository,
	hasher domain.PasswordHasher,
	clock domain.Clock,
	eventBus eventbus.Bus,
) SubmitClassifiedAdCommand {
	return func(args SubmitClassifiedAdCommandArgs) (string, error) {
		email, err := domain.NewEmail(args.SellerEmail)
		if err != nil {
			return "", err
		}

		password, err := domain.NewPassword(args.SellerPassword, hasher)
		if err != nil {
			return "", err
		}

		seller, err := domain.NewSeller(email, args.SellerPseudo, password)
		if err != nil {
			return "", err
		}

		category, err := domain.NewCategory(args.Category)
		if err != nil {
			return "", err
		}

		location, err := domain.NewLocation(args.ZipCode, args.CityName)
		if err != nil {
			return "", err
		}

		submissionDate := domain.NewSubmissionDate(clock.Now())

		ad, err := domain.NewClassifiedAd(
			args.Title,
			args.Description,
			args.PriceInCents,
			seller,
			args.ImageURLs,
			category,
			location,
			submissionDate,
		)
		if err != nil {
			return "", err
		}

		if err := repo.Save(ad); err != nil {
			return "", err
		}

		event := domain.NewClassifiedAdSubmittedEventFromClassifiedAd(ad)
		if err := eventBus.Publish(event); err != nil {
			return "", err
		}

		return ad.ID().String(), nil
	}
}

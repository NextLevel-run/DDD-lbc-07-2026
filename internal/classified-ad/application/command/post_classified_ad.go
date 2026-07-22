package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// PostClassifiedAdCommandArgs contains input data for the command
type PostClassifiedAdCommandArgs struct {
	SellerId      string
	Title         string
	Description   string
	PriceAmount   int64
	PriceCurrency string
	Category      string
	PhotoURLs     []string
}

// PostClassifiedAdCommand is the command function type
type PostClassifiedAdCommand func(args PostClassifiedAdCommandArgs) (string, error)

// BuildPostClassifiedAdCommand builds a command with dependencies injected
func BuildPostClassifiedAdCommand(repo domain.ClassifiedAdRepository, eventBus eventbus.Bus) PostClassifiedAdCommand {
	return func(args PostClassifiedAdCommandArgs) (string, error) {
		classifiedAd, err := domain.NewClassifiedAd(
			args.SellerId,
			args.Title,
			args.Description,
			args.PriceAmount,
			args.PriceCurrency,
			args.Category,
			args.PhotoURLs,
		)
		if err != nil {
			return "", err
		}

		if err := repo.Save(classifiedAd); err != nil {
			return "", err
		}

		event := domain.NewClassifiedAdPostedEventFrom(classifiedAd)
		if err := eventBus.Publish(event); err != nil {
			return "", err
		}

		return classifiedAd.ID(), nil
	}
}

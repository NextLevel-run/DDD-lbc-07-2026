package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// MakeOfferCommandArgs contains input data for the command.
type MakeOfferCommandArgs struct {
	AdID          string
	BuyerEmail    string
	BuyerPseudo   string
	AmountInCents int64
	Message       string
}

// MakeOfferCommand is the command function type.
type MakeOfferCommand func(args MakeOfferCommandArgs) error

// BuildMakeOfferCommand builds a command with dependencies injected.
func BuildMakeOfferCommand(
	repo domain.ClassifiedAdRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) MakeOfferCommand {
	return func(args MakeOfferCommandArgs) error {
		adID, err := uuid.Parse(args.AdID)
		if err != nil {
			return domain.ErrClassifiedAdNotFound
		}

		ad, err := repo.FindByID(adID)
		if err != nil {
			return err
		}

		if !ad.CanReceiveOffer() {
			return domain.ErrAdNotAvailable
		}

		buyerEmail, err := domain.NewEmail(args.BuyerEmail)
		if err != nil {
			return err
		}

		if err := domain.ValidateOfferMessage(args.Message); err != nil {
			return err
		}

		if err := domain.ValidateOfferAmount(args.AmountInCents); err != nil {
			return err
		}

		event := domain.NewBuyerOfferMadeEvent(ad, buyerEmail.String(), args.BuyerPseudo, args.AmountInCents, args.Message, clock.Now())
		return eventBus.Publish(event)
	}
}

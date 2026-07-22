package command

import (
	"strings"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// PostClassifiedAdCommandArgs contains the input data for posting an ad.
// SellerPassword is the plaintext password: it is hashed here and never leaves
// this layer in clear text.
type PostClassifiedAdCommandArgs struct {
	Title          string
	Description    string
	PriceInCents   int64
	Category       string
	PhotoURLs      []string
	Location       string
	SellerEmail    string
	SellerNickname string
	SellerPassword string
}

// PostClassifiedAdCommand posts a new ad and returns its ID.
type PostClassifiedAdCommand func(args PostClassifiedAdCommandArgs) (string, error)

// BuildPostClassifiedAdCommand wires the command with its dependencies.
func BuildPostClassifiedAdCommand(
	repo domain.ClassifiedAdRepository,
	hasher domain.PasswordHasher,
	eventBus eventbus.Bus,
) PostClassifiedAdCommand {
	return func(args PostClassifiedAdCommandArgs) (string, error) {
		// 1. Guard the plaintext password before hashing (the domain never sees it).
		if strings.TrimSpace(args.SellerPassword) == "" {
			return "", domain.ErrEmptyPassword
		}

		// 2. Hash the password so only the hash reaches the domain.
		hashedPassword, err := hasher.Hash(args.SellerPassword)
		if err != nil {
			return "", err
		}

		// 3. Build the aggregate (validates all invariants).
		ad, err := domain.NewClassifiedAd(
			args.Title,
			args.Description,
			args.PriceInCents,
			args.Category,
			args.PhotoURLs,
			args.Location,
			args.SellerEmail,
			args.SellerNickname,
			hashedPassword,
		)
		if err != nil {
			return "", err
		}

		// 4. Persist.
		if err := repo.Save(ad); err != nil {
			return "", err
		}

		// 5. Emit the domain event AFTER successful persistence.
		event := domain.NewClassifiedAdPostedEventFrom(ad)
		if err := eventBus.Publish(event); err != nil {
			return "", err
		}

		return ad.ID(), nil
	}
}

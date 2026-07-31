package command

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// ExpireOutdatedAdsCommand is the command function type.
type ExpireOutdatedAdsCommand func() (int, error)

// BuildExpireOutdatedAdsCommand builds a command with dependencies injected.
func BuildExpireOutdatedAdsCommand(
	repo domain.ClassifiedAdRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) ExpireOutdatedAdsCommand {
	return func() (int, error) {
		now := clock.Now()

		ads, err := repo.FindExpirable(now)
		if err != nil {
			return 0, err
		}

		count := 0
		for _, ad := range ads {
			if !ad.Expire(now) {
				continue
			}

			if err := repo.Save(ad); err != nil {
				return count, err
			}

			event := domain.NewClassifiedAdExpiredEventFromClassifiedAd(ad)
			if err := eventBus.Publish(event); err != nil {
				return count, err
			}

			count++
		}

		return count, nil
	}
}

package query

import (
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

// SearchClassifiedAdsQueryArgs holds the filters, sorting and pagination for a search.
type SearchClassifiedAdsQueryArgs struct {
	Category        *string
	ZipCode         *string
	CityName        *string
	Keywords        *string
	MinPriceInCents *int64
	MaxPriceInCents *int64
	SortBy          string
	Limit           int
	Offset          int
}

// ClassifiedAdListItemView is the read model for a single item in a search result list.
type ClassifiedAdListItemView struct {
	ID             string
	Title          string
	PriceInCents   int64
	Category       string
	CityName       string
	ZipCode        string
	FirstImageURL  string
	SubmissionDate time.Time
}

// SearchClassifiedAdsQuery searches for online classified ads matching the given criteria.
type SearchClassifiedAdsQuery func(args SearchClassifiedAdsQueryArgs) ([]ClassifiedAdListItemView, error)

// BuildSearchClassifiedAdsQuery wires the SearchClassifiedAdsQuery use case.
func BuildSearchClassifiedAdsQuery(repo domain.ClassifiedAdRepository) SearchClassifiedAdsQuery {
	return func(args SearchClassifiedAdsQueryArgs) ([]ClassifiedAdListItemView, error) {
		sortBy := args.SortBy
		if sortBy == "" {
			sortBy = "date_desc"
		}

		limit := args.Limit
		if limit <= 0 {
			limit = 20
		}

		offset := args.Offset
		if offset < 0 {
			offset = 0
		}

		criteria := domain.SearchCriteria{
			ZipCode:         args.ZipCode,
			CityName:        args.CityName,
			Keywords:        args.Keywords,
			MinPriceInCents: args.MinPriceInCents,
			MaxPriceInCents: args.MaxPriceInCents,
			// OnlineOnly is a fixed business rule of this query, not a client-provided
			// filter: search must never surface deleted or expired ads.
			OnlineOnly: true,
			SortBy:     sortBy,
			Limit:      limit,
			Offset:     offset,
		}

		if args.Category != nil {
			category, err := domain.NewCategory(*args.Category)
			if err != nil {
				return nil, err
			}
			criteria.Category = &category
		}

		ads, err := repo.Search(criteria)
		if err != nil {
			return nil, err
		}

		views := make([]ClassifiedAdListItemView, 0, len(ads))
		for _, ad := range ads {
			firstImageURL := ""
			if imageURLs := ad.ImageURLs(); len(imageURLs) > 0 {
				firstImageURL = imageURLs[0]
			}

			views = append(views, ClassifiedAdListItemView{
				ID:             ad.ID().String(),
				Title:          ad.Title(),
				PriceInCents:   ad.Price().AmountInCents(),
				Category:       string(ad.Category()),
				CityName:       ad.Location().CityName(),
				ZipCode:        ad.Location().ZipCode(),
				FirstImageURL:  firstImageURL,
				SubmissionDate: ad.SubmissionDate().Time(),
			})
		}

		return views, nil
	}
}

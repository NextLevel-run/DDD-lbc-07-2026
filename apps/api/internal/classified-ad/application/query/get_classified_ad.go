package query

import (
	"time"

	"github.com/google/uuid"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

// ClassifiedAdView is the read model for a single classified ad detail.
type ClassifiedAdView struct {
	ID             string
	Title          string
	Description    string
	PriceInCents   int64
	Category       string
	SellerPseudo   string
	ImageURLs      []string
	ZipCode        string
	CityName       string
	SubmissionDate time.Time
}

// GetClassifiedAdQuery retrieves the detail view of an online classified ad by id.
type GetClassifiedAdQuery func(id string) (ClassifiedAdView, error)

// BuildGetClassifiedAdQuery wires the GetClassifiedAdQuery use case.
func BuildGetClassifiedAdQuery(repo domain.ClassifiedAdRepository) GetClassifiedAdQuery {
	return func(id string) (ClassifiedAdView, error) {
		adID, err := uuid.Parse(id)
		if err != nil {
			return ClassifiedAdView{}, domain.ErrClassifiedAdNotFound
		}

		ad, err := repo.FindByID(adID)
		if err != nil {
			return ClassifiedAdView{}, domain.ErrClassifiedAdNotFound
		}

		if !ad.IsOnline() {
			return ClassifiedAdView{}, domain.ErrClassifiedAdNotFound
		}

		return ClassifiedAdView{
			ID:             ad.ID().String(),
			Title:          ad.Title(),
			Description:    ad.Description(),
			PriceInCents:   ad.Price().AmountInCents(),
			Category:       string(ad.Category()),
			SellerPseudo:   ad.Seller().Pseudo(),
			ImageURLs:      ad.ImageURLs(),
			ZipCode:        ad.Location().ZipCode(),
			CityName:       ad.Location().CityName(),
			SubmissionDate: ad.SubmissionDate().Time(),
		}, nil
	}
}

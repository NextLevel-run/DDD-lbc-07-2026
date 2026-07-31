package query

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

// fakeClassifiedAdRepository is a minimal, test-only in-memory implementation of
// domain.ClassifiedAdRepository, used until a real adapter/driven/inmemory exists.
type fakeClassifiedAdRepository struct {
	mu  sync.RWMutex
	ads map[uuid.UUID]*domain.ClassifiedAd
}

func newFakeClassifiedAdRepository() *fakeClassifiedAdRepository {
	return &fakeClassifiedAdRepository{
		ads: make(map[uuid.UUID]*domain.ClassifiedAd),
	}
}

func (r *fakeClassifiedAdRepository) Save(ad *domain.ClassifiedAd) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ads[ad.ID()] = ad
	return nil
}

func (r *fakeClassifiedAdRepository) FindByID(id uuid.UUID) (*domain.ClassifiedAd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ad, ok := r.ads[id]
	if !ok {
		return nil, domain.ErrClassifiedAdNotFound
	}
	return ad, nil
}

func (r *fakeClassifiedAdRepository) FindExpirable(now time.Time) ([]*domain.ClassifiedAd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.ClassifiedAd
	for _, ad := range r.ads {
		if ad.IsExpirable(now) {
			result = append(result, ad)
		}
	}
	return result, nil
}

func (r *fakeClassifiedAdRepository) Search(criteria domain.SearchCriteria) ([]*domain.ClassifiedAd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.ClassifiedAd
	for _, ad := range r.ads {
		if criteria.OnlineOnly && !ad.IsOnline() {
			continue
		}
		if criteria.Category != nil && ad.Category() != *criteria.Category {
			continue
		}
		if criteria.ZipCode != nil && ad.Location().ZipCode() != *criteria.ZipCode {
			continue
		}
		if criteria.CityName != nil && ad.Location().CityName() != *criteria.CityName {
			continue
		}
		if criteria.MinPriceInCents != nil && ad.Price().AmountInCents() < *criteria.MinPriceInCents {
			continue
		}
		if criteria.MaxPriceInCents != nil && ad.Price().AmountInCents() > *criteria.MaxPriceInCents {
			continue
		}
		if criteria.Keywords != nil {
			keywords := strings.ToLower(*criteria.Keywords)
			title := strings.ToLower(ad.Title())
			description := strings.ToLower(ad.Description())
			if !strings.Contains(title, keywords) && !strings.Contains(description, keywords) {
				continue
			}
		}
		result = append(result, ad)
	}

	sortBy := criteria.SortBy
	if sortBy == "" {
		sortBy = "date_desc"
	}

	sortAds(result, sortBy)

	offset := criteria.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(result) {
		return []*domain.ClassifiedAd{}, nil
	}
	result = result[offset:]

	limit := criteria.Limit
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	return result, nil
}

func sortAds(ads []*domain.ClassifiedAd, sortBy string) {
	for i := 1; i < len(ads); i++ {
		for j := i; j > 0 && adsLess(ads[j], ads[j-1], sortBy); j-- {
			ads[j], ads[j-1] = ads[j-1], ads[j]
		}
	}
}

func adsLess(a, b *domain.ClassifiedAd, sortBy string) bool {
	switch sortBy {
	case "date_asc":
		return a.SubmissionDate().Time().Before(b.SubmissionDate().Time())
	case "price_asc":
		return a.Price().AmountInCents() < b.Price().AmountInCents()
	case "price_desc":
		return a.Price().AmountInCents() > b.Price().AmountInCents()
	default: // date_desc
		return a.SubmissionDate().Time().After(b.SubmissionDate().Time())
	}
}

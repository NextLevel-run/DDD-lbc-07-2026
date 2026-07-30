package inmemory

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

// InMemoryClassifiedAdRepository is a thread-safe in-memory implementation of
// domain.ClassifiedAdRepository, intended for tests and local development.
type InMemoryClassifiedAdRepository struct {
	mu  sync.RWMutex
	ads map[uuid.UUID]*domain.ClassifiedAd
}

// NewInMemoryClassifiedAdRepository builds an empty InMemoryClassifiedAdRepository.
func NewInMemoryClassifiedAdRepository() *InMemoryClassifiedAdRepository {
	return &InMemoryClassifiedAdRepository{
		ads: make(map[uuid.UUID]*domain.ClassifiedAd),
	}
}

// Save inserts or updates the given classified ad.
func (r *InMemoryClassifiedAdRepository) Save(ad *domain.ClassifiedAd) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ads[ad.ID()] = ad
	return nil
}

// FindByID returns the classified ad with the given id, or ErrClassifiedAdNotFound.
func (r *InMemoryClassifiedAdRepository) FindByID(id uuid.UUID) (*domain.ClassifiedAd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ad, ok := r.ads[id]
	if !ok {
		return nil, domain.ErrClassifiedAdNotFound
	}
	return ad, nil
}

// FindExpirable returns all published ads whose lifetime has elapsed as of now.
func (r *InMemoryClassifiedAdRepository) FindExpirable(now time.Time) ([]*domain.ClassifiedAd, error) {
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

// Search returns online ads matching the given criteria, sorted and paginated.
func (r *InMemoryClassifiedAdRepository) Search(criteria domain.SearchCriteria) ([]*domain.ClassifiedAd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.ClassifiedAd
	for _, ad := range r.ads {
		if !ad.IsOnline() {
			continue
		}
		if !matchesCriteria(ad, criteria) {
			continue
		}
		result = append(result, ad)
	}

	sortAds(result, criteria.SortBy)

	return paginate(result, criteria.Limit, criteria.Offset), nil
}

func matchesCriteria(ad *domain.ClassifiedAd, criteria domain.SearchCriteria) bool {
	if criteria.Category != nil && ad.Category() != *criteria.Category {
		return false
	}
	if criteria.ZipCode != nil && ad.Location().ZipCode() != *criteria.ZipCode {
		return false
	}
	if criteria.CityName != nil && ad.Location().CityName() != *criteria.CityName {
		return false
	}
	if criteria.MinPriceInCents != nil && ad.Price().AmountInCents() < *criteria.MinPriceInCents {
		return false
	}
	if criteria.MaxPriceInCents != nil && ad.Price().AmountInCents() > *criteria.MaxPriceInCents {
		return false
	}
	if criteria.Keywords != nil {
		keywords := strings.ToLower(*criteria.Keywords)
		title := strings.ToLower(ad.Title())
		description := strings.ToLower(ad.Description())
		if !strings.Contains(title, keywords) && !strings.Contains(description, keywords) {
			return false
		}
	}
	return true
}

func sortAds(ads []*domain.ClassifiedAd, sortBy string) {
	switch sortBy {
	case "date_asc":
		sort.SliceStable(ads, func(i, j int) bool {
			return ads[i].PublishedAt().Before(ads[j].PublishedAt())
		})
	case "price_asc":
		sort.SliceStable(ads, func(i, j int) bool {
			return ads[i].Price().AmountInCents() < ads[j].Price().AmountInCents()
		})
	case "price_desc":
		sort.SliceStable(ads, func(i, j int) bool {
			return ads[i].Price().AmountInCents() > ads[j].Price().AmountInCents()
		})
	default: // "date_desc" and unknown values
		sort.SliceStable(ads, func(i, j int) bool {
			return ads[i].PublishedAt().After(ads[j].PublishedAt())
		})
	}
}

func paginate(ads []*domain.ClassifiedAd, limit, offset int) []*domain.ClassifiedAd {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(ads) {
		return []*domain.ClassifiedAd{}
	}
	ads = ads[offset:]
	if limit > 0 && limit < len(ads) {
		ads = ads[:limit]
	}
	return ads
}

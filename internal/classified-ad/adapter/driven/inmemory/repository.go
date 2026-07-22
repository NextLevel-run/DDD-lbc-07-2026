package inmemory

import (
	"errors"
	"sync"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

// InMemoryClassifiedAdRepository is a thread-safe in-memory implementation of
// the ClassifiedAdRepository port, used for tests and local development.
type InMemoryClassifiedAdRepository struct {
	mu  sync.RWMutex
	ads map[string]*domain.ClassifiedAd
}

// compile-time check that the adapter satisfies the domain port.
var _ domain.ClassifiedAdRepository = (*InMemoryClassifiedAdRepository)(nil)

func NewInMemoryClassifiedAdRepository() *InMemoryClassifiedAdRepository {
	return &InMemoryClassifiedAdRepository{
		ads: make(map[string]*domain.ClassifiedAd),
	}
}

func (r *InMemoryClassifiedAdRepository) Save(ad *domain.ClassifiedAd) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ads[ad.ID()] = ad
	return nil
}

func (r *InMemoryClassifiedAdRepository) GetById(id string) (*domain.ClassifiedAd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ad, exists := r.ads[id]
	if !exists {
		return nil, errors.New("classified ad not found")
	}
	return ad, nil
}

func (r *InMemoryClassifiedAdRepository) FindAll(filters domain.FindAllFilters) ([]*domain.ClassifiedAd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.ClassifiedAd, 0, len(r.ads))
	for _, ad := range r.ads {
		if filters.Status != nil && ad.Status() != *filters.Status {
			continue
		}
		if filters.Category != nil && ad.Category() != *filters.Category {
			continue
		}
		result = append(result, ad)
	}
	return result, nil
}

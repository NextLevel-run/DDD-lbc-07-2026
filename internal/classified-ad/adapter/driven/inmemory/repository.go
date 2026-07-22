package inmemory

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"sync"
)

// InMemoryClassifiedAdRepository is a thread-safe in-memory implementation
type InMemoryClassifiedAdRepository struct {
	classifiedAds map[string]*domain.ClassifiedAd
	mu            sync.RWMutex
}

func NewInMemoryClassifiedAdRepository() *InMemoryClassifiedAdRepository {
	return &InMemoryClassifiedAdRepository{
		classifiedAds: make(map[string]*domain.ClassifiedAd),
	}
}

func (r *InMemoryClassifiedAdRepository) Save(classifiedAd *domain.ClassifiedAd) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.classifiedAds[classifiedAd.ID()] = classifiedAd
	return nil
}

func (r *InMemoryClassifiedAdRepository) GetById(id string) (*domain.ClassifiedAd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	classifiedAd, exists := r.classifiedAds[id]
	if !exists {
		return nil, domain.ErrClassifiedAdNotFound
	}
	return classifiedAd, nil
}

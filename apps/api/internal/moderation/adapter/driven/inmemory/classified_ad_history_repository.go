package inmemory

import (
	"sync"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// InMemoryClassifiedAdHistoryRepository is a thread-safe in-memory implementation
// of domain.ClassifiedAdHistoryRepository, keyed by classified ad ID, intended
// for tests and local development.
type InMemoryClassifiedAdHistoryRepository struct {
	mu        sync.RWMutex
	histories map[string]*domain.ClassifiedAdHistory
}

// NewInMemoryClassifiedAdHistoryRepository builds an empty InMemoryClassifiedAdHistoryRepository.
func NewInMemoryClassifiedAdHistoryRepository() *InMemoryClassifiedAdHistoryRepository {
	return &InMemoryClassifiedAdHistoryRepository{
		histories: make(map[string]*domain.ClassifiedAdHistory),
	}
}

// Save inserts or updates the history of a classified ad.
func (r *InMemoryClassifiedAdHistoryRepository) Save(history *domain.ClassifiedAdHistory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.histories[history.ClassifiedAdID()] = history
	return nil
}

// FindByClassifiedAdID returns the history of the given classified ad, or
// ErrClassifiedAdHistoryNotFound when no history exists yet.
func (r *InMemoryClassifiedAdHistoryRepository) FindByClassifiedAdID(classifiedAdID string) (*domain.ClassifiedAdHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	history, ok := r.histories[classifiedAdID]
	if !ok {
		return nil, domain.ErrClassifiedAdHistoryNotFound
	}
	return history, nil
}

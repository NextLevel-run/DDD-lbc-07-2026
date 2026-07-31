package inmemory

import (
	"sync"

	"github.com/google/uuid"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// InMemoryModeratorRepository is a thread-safe in-memory implementation of
// domain.ModeratorRepository, intended for tests and local development.
type InMemoryModeratorRepository struct {
	mu         sync.RWMutex
	moderators map[uuid.UUID]*domain.Moderator
}

// NewInMemoryModeratorRepository builds an empty InMemoryModeratorRepository.
func NewInMemoryModeratorRepository() *InMemoryModeratorRepository {
	return &InMemoryModeratorRepository{
		moderators: make(map[uuid.UUID]*domain.Moderator),
	}
}

// Save inserts or updates the given moderator.
func (r *InMemoryModeratorRepository) Save(moderator *domain.Moderator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.moderators[moderator.ID()] = moderator
	return nil
}

// FindByID returns the moderator with the given id, or ErrModeratorNotFound.
func (r *InMemoryModeratorRepository) FindByID(id uuid.UUID) (*domain.Moderator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	moderator, ok := r.moderators[id]
	if !ok {
		return nil, domain.ErrModeratorNotFound
	}
	return moderator, nil
}

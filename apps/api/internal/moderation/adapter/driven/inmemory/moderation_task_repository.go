package inmemory

import (
	"sync"

	"github.com/google/uuid"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// InMemoryModerationTaskRepository is a thread-safe in-memory implementation of
// domain.ModerationTaskRepository, intended for tests and local development.
type InMemoryModerationTaskRepository struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]*domain.ModerationTask
}

// NewInMemoryModerationTaskRepository builds an empty InMemoryModerationTaskRepository.
func NewInMemoryModerationTaskRepository() *InMemoryModerationTaskRepository {
	return &InMemoryModerationTaskRepository{
		tasks: make(map[uuid.UUID]*domain.ModerationTask),
	}
}

// Save inserts or updates the given moderation task.
func (r *InMemoryModerationTaskRepository) Save(task *domain.ModerationTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[task.ID()] = task
	return nil
}

// FindByID returns the task with the given id, or ErrModerationTaskNotFound.
func (r *InMemoryModerationTaskRepository) FindByID(id uuid.UUID) (*domain.ModerationTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, domain.ErrModerationTaskNotFound
	}
	return task, nil
}

// FindAll returns every active task (pending and claimed).
func (r *InMemoryModerationTaskRepository) FindAll() ([]*domain.ModerationTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]*domain.ModerationTask, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// Delete physically removes the task with the given id, or returns
// ErrModerationTaskNotFound when no task matches.
func (r *InMemoryModerationTaskRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return domain.ErrModerationTaskNotFound
	}
	delete(r.tasks, id)
	return nil
}

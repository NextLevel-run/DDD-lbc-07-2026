package domain

import (
	"time"

	"github.com/google/uuid"
)

// Clock is a port abstracting the current time.
type Clock interface {
	Now() time.Time
}

// ModerationTaskRepository is the persistence port for ModerationTask aggregates.
type ModerationTaskRepository interface {
	Save(task *ModerationTask) error
	// FindByID returns ErrModerationTaskNotFound when no task matches.
	FindByID(id uuid.UUID) (*ModerationTask, error)
	// FindAll returns every active task (pending and claimed).
	FindAll() ([]*ModerationTask, error)
	// Delete physically removes a completed task. Returns ErrModerationTaskNotFound
	// when no task matches.
	Delete(id uuid.UUID) error
}

// ModeratorRepository is the persistence port for Moderator aggregates.
type ModeratorRepository interface {
	Save(moderator *Moderator) error
	// FindByID returns ErrModeratorNotFound when no moderator matches.
	FindByID(id uuid.UUID) (*Moderator, error)
}

// ClassifiedAdHistoryRepository is the persistence port for ClassifiedAdHistory
// aggregates, keyed by the classified ad ID.
type ClassifiedAdHistoryRepository interface {
	Save(history *ClassifiedAdHistory) error
	// FindByClassifiedAdID returns ErrClassifiedAdHistoryNotFound when no history
	// exists yet for the given ad.
	FindByClassifiedAdID(classifiedAdID string) (*ClassifiedAdHistory, error)
}

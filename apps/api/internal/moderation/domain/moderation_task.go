package domain

import (
	"time"

	"github.com/google/uuid"
)

// ModerationTask is the aggregate root representing a classified ad awaiting
// moderation. Tasks live in a shared queue visible to all moderators; a task
// must be claimed (exclusive lock, no timeout) before being completed, and is
// physically deleted after completion — the audit trail is carried by
// ClassifiedAdHistory, not by the task.
type ModerationTask struct {
	id             uuid.UUID
	createdAt      time.Time
	classifiedAdID string
	moderatorID    *uuid.UUID // nil while the task is unclaimed
	claimedAt      *time.Time // nil while the task is unclaimed
}

// NewModerationTask validates and builds a new ModerationTask for a classified ad.
// Each submission or re-submission of an ad creates a new task with a new ID.
func NewModerationTask(classifiedAdID string, createdAt time.Time) (*ModerationTask, error) {
	if classifiedAdID == "" {
		return nil, ErrEmptyClassifiedAdID
	}
	return &ModerationTask{
		id:             uuid.New(),
		createdAt:      createdAt,
		classifiedAdID: classifiedAdID,
	}, nil
}

// ID returns the aggregate identifier.
func (t *ModerationTask) ID() uuid.UUID {
	return t.id
}

// CreatedAt returns when the task was created.
func (t *ModerationTask) CreatedAt() time.Time {
	return t.createdAt
}

// ClassifiedAdID returns the ID of the classified ad under moderation.
func (t *ModerationTask) ClassifiedAdID() string {
	return t.classifiedAdID
}

// ModeratorID returns the ID of the moderator holding the claim, or nil if unclaimed.
func (t *ModerationTask) ModeratorID() *uuid.UUID {
	return t.moderatorID
}

// ClaimedAt returns when the task was claimed, or nil if unclaimed.
func (t *ModerationTask) ClaimedAt() *time.Time {
	return t.claimedAt
}

// IsClaimed returns true if a moderator currently holds the task.
func (t *ModerationTask) IsClaimed() bool {
	return t.moderatorID != nil
}

// Claim locks the task exclusively for the given moderator. The lock has no
// timeout. Returns ErrTaskAlreadyClaimed if any moderator (including the same
// one) already holds the task.
func (t *ModerationTask) Claim(moderatorID uuid.UUID, now time.Time) error {
	if t.IsClaimed() {
		return ErrTaskAlreadyClaimed
	}
	t.moderatorID = &moderatorID
	t.claimedAt = &now
	return nil
}

// Complete verifies that the given moderator is allowed to complete the task:
// only the moderator holding the claim may complete it. Returns ErrNotTaskOwner
// otherwise (including when the task was never claimed — nobody owns it).
// The physical deletion of the completed task is handled by the caller.
func (t *ModerationTask) Complete(moderatorID uuid.UUID) error {
	if t.moderatorID == nil || *t.moderatorID != moderatorID {
		return ErrNotTaskOwner
	}
	return nil
}

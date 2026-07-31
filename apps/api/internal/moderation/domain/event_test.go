package domain_test

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newClaimedTask(t *testing.T, createdAt time.Time, moderatorID uuid.UUID, claimedAt time.Time) *domain.ModerationTask {
	t.Helper()
	task := newValidTask(t, createdAt)
	require.NoError(t, task.Claim(moderatorID, claimedAt))
	return task
}

func TestNewModerationTaskClaimedEventFromTask(t *testing.T) {
	now := time.Now()
	moderatorID := uuid.New()
	claimedAt := now.Add(time.Minute)
	task := newClaimedTask(t, now, moderatorID, claimedAt)

	event := domain.NewModerationTaskClaimedEventFromTask(task)

	assert.Equal(t, "ModerationTaskClaimed", event.EventType())
	assert.Equal(t, task.ID().String(), event.TaskID)
	assert.Equal(t, task.ClassifiedAdID(), event.ClassifiedAdID)
	assert.Equal(t, moderatorID.String(), event.ModeratorID)
	assert.Equal(t, claimedAt, event.OccurredAt)
}

func TestNewModerationTaskCompletedEventFromTask(t *testing.T) {
	now := time.Now()
	moderatorID := uuid.New()
	task := newClaimedTask(t, now, moderatorID, now)
	occurredAt := now.Add(time.Hour)

	event := domain.NewModerationTaskCompletedEventFromTask(task, occurredAt)

	assert.Equal(t, "ModerationTaskCompleted", event.EventType())
	assert.Equal(t, task.ID().String(), event.TaskID)
	assert.Equal(t, task.ClassifiedAdID(), event.ClassifiedAdID)
	assert.Equal(t, moderatorID.String(), event.ModeratorID)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

func TestNewClassifiedAdApprovedEventFromTask(t *testing.T) {
	now := time.Now()
	moderatorID := uuid.New()
	task := newClaimedTask(t, now, moderatorID, now)
	occurredAt := now.Add(time.Hour)

	event := domain.NewClassifiedAdApprovedEventFromTask(task, occurredAt)

	assert.Equal(t, "ClassifiedAdApproved", event.EventType())
	assert.Equal(t, task.ClassifiedAdID(), event.ClassifiedAdID)
	assert.Equal(t, moderatorID.String(), event.ModeratorID)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

func TestNewClassifiedAdRejectedEventFromTask(t *testing.T) {
	now := time.Now()
	moderatorID := uuid.New()
	task := newClaimedTask(t, now, moderatorID, now)
	occurredAt := now.Add(time.Hour)

	event := domain.NewClassifiedAdRejectedEventFromTask(task, domain.RejectReasonSuspectPrice, occurredAt)

	assert.Equal(t, "ClassifiedAdRejected", event.EventType())
	assert.Equal(t, task.ClassifiedAdID(), event.ClassifiedAdID)
	assert.Equal(t, moderatorID.String(), event.ModeratorID)
	assert.Equal(t, string(domain.RejectReasonSuspectPrice), event.Reason)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

func TestNewClassifiedAdChallengedEventFromTask(t *testing.T) {
	now := time.Now()
	moderatorID := uuid.New()
	task := newClaimedTask(t, now, moderatorID, now)
	occurredAt := now.Add(time.Hour)

	event := domain.NewClassifiedAdChallengedEventFromTask(task, domain.ChallengeReasonPriceToVerify, occurredAt)

	assert.Equal(t, "ClassifiedAdChallenged", event.EventType())
	assert.Equal(t, task.ClassifiedAdID(), event.ClassifiedAdID)
	assert.Equal(t, moderatorID.String(), event.ModeratorID)
	assert.Equal(t, string(domain.ChallengeReasonPriceToVerify), event.Reason)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

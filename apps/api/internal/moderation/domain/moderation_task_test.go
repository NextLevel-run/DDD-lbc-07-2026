package domain_test

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newValidTask(t *testing.T, createdAt time.Time) *domain.ModerationTask {
	t.Helper()
	task, err := domain.NewModerationTask("ad-123", createdAt)
	require.NoError(t, err)
	return task
}

func TestNewModerationTask(t *testing.T) {
	now := time.Now()

	t.Run("valid task is created unclaimed", func(t *testing.T) {
		task, err := domain.NewModerationTask("ad-123", now)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, task.ID())
		assert.Equal(t, now, task.CreatedAt())
		assert.Equal(t, "ad-123", task.ClassifiedAdID())
		assert.Nil(t, task.ModeratorID())
		assert.Nil(t, task.ClaimedAt())
		assert.False(t, task.IsClaimed())
	})

	t.Run("each task gets a new id even for the same ad", func(t *testing.T) {
		first, err := domain.NewModerationTask("ad-123", now)
		require.NoError(t, err)
		second, err := domain.NewModerationTask("ad-123", now)
		require.NoError(t, err)
		assert.NotEqual(t, first.ID(), second.ID())
	})

	t.Run("empty classified ad id is rejected", func(t *testing.T) {
		_, err := domain.NewModerationTask("", now)
		assert.ErrorIs(t, err, domain.ErrEmptyClassifiedAdID)
	})
}

func TestModerationTask_Claim(t *testing.T) {
	now := time.Now()
	moderatorID := uuid.New()

	t.Run("unclaimed task can be claimed", func(t *testing.T) {
		task := newValidTask(t, now)
		claimedAt := now.Add(time.Minute)

		err := task.Claim(moderatorID, claimedAt)

		require.NoError(t, err)
		assert.True(t, task.IsClaimed())
		require.NotNil(t, task.ModeratorID())
		assert.Equal(t, moderatorID, *task.ModeratorID())
		require.NotNil(t, task.ClaimedAt())
		assert.Equal(t, claimedAt, *task.ClaimedAt())
	})

	t.Run("claiming an already claimed task fails", func(t *testing.T) {
		task := newValidTask(t, now)
		require.NoError(t, task.Claim(moderatorID, now))

		err := task.Claim(uuid.New(), now.Add(time.Minute))

		assert.ErrorIs(t, err, domain.ErrTaskAlreadyClaimed)
		assert.Equal(t, moderatorID, *task.ModeratorID())
	})

	t.Run("claiming twice even by the same moderator fails", func(t *testing.T) {
		task := newValidTask(t, now)
		require.NoError(t, task.Claim(moderatorID, now))

		err := task.Claim(moderatorID, now.Add(time.Minute))

		assert.ErrorIs(t, err, domain.ErrTaskAlreadyClaimed)
	})
}

func TestModerationTask_Complete(t *testing.T) {
	now := time.Now()
	ownerID := uuid.New()

	t.Run("owner can complete the task", func(t *testing.T) {
		task := newValidTask(t, now)
		require.NoError(t, task.Claim(ownerID, now))

		assert.NoError(t, task.Complete(ownerID))
	})

	t.Run("another moderator cannot complete the task", func(t *testing.T) {
		task := newValidTask(t, now)
		require.NoError(t, task.Claim(ownerID, now))

		err := task.Complete(uuid.New())

		assert.ErrorIs(t, err, domain.ErrNotTaskOwner)
	})

	t.Run("an unclaimed task cannot be completed", func(t *testing.T) {
		task := newValidTask(t, now)

		err := task.Complete(ownerID)

		assert.ErrorIs(t, err, domain.ErrNotTaskOwner)
	})
}

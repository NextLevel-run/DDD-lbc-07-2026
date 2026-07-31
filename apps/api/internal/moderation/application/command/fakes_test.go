package command

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"
)

// fakeClock is a settable implementation of domain.Clock for deterministic tests.
type fakeClock struct {
	now time.Time
}

func newFakeClock(t time.Time) *fakeClock {
	return &fakeClock{now: t}
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

// seedModerator creates and stores a moderator with the given full name.
func seedModerator(t *testing.T, repo *inmemory.InMemoryModeratorRepository, fullName string) *domain.Moderator {
	t.Helper()

	moderator, err := domain.NewModerator(fullName)
	require.NoError(t, err)
	require.NoError(t, repo.Save(moderator))
	return moderator
}

// seedTask creates and stores an unclaimed moderation task for the given ad.
func seedTask(t *testing.T, repo *inmemory.InMemoryModerationTaskRepository, classifiedAdID string, createdAt time.Time) *domain.ModerationTask {
	t.Helper()

	task, err := domain.NewModerationTask(classifiedAdID, createdAt)
	require.NoError(t, err)
	require.NoError(t, repo.Save(task))
	return task
}

// seedClaimedTask creates and stores a task already claimed by the given moderator.
func seedClaimedTask(t *testing.T, repo *inmemory.InMemoryModerationTaskRepository, classifiedAdID string, moderator *domain.Moderator, now time.Time) *domain.ModerationTask {
	t.Helper()

	task := seedTask(t, repo, classifiedAdID, now)
	require.NoError(t, task.Claim(moderator.ID(), now))
	require.NoError(t, repo.Save(task))
	return task
}

func assertNoEventsEmitted(t *testing.T, collector *eventbustesting.EventCollector) {
	t.Helper()
	assert.Empty(t, collector.GetEvents(), "Expected no events to be emitted")
}

// assertTaskStillStored verifies that the task still exists in the repository.
func assertTaskStillStored(t *testing.T, repo *inmemory.InMemoryModerationTaskRepository, task *domain.ModerationTask) {
	t.Helper()
	_, err := repo.FindByID(task.ID())
	assert.NoError(t, err, "Expected task to still be stored")
}

package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

func TestListModerationTasksQuery_EmptyQueue(t *testing.T) {
	// Given
	setup := setupQueryTest(t)
	query := BuildListModerationTasksQuery(setup.taskRepo, setup.moderatorRepo, setup.historyRepo)

	// When
	items, err := query()

	// Then
	require.NoError(t, err)
	assert.Empty(t, items, "Expected an empty list when no task exists")
}

func TestListModerationTasksQuery_PendingAndClaimedTasks(t *testing.T) {
	// Given: an older pending task and a newer claimed one
	setup := setupQueryTest(t)
	now := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)

	pendingTask := seedTask(t, setup, "ad-1", now.Add(-2*time.Hour))
	seedSubmittedHistory(t, setup, "ad-1", "Vélo hollandais", now.Add(-2*time.Hour))

	moderator := seedModerator(t, setup, "Jane Doe")
	claimedTask := seedClaimedTask(t, setup, "ad-2", moderator, now.Add(-time.Hour), now)
	seedSubmittedHistory(t, setup, "ad-2", "Table basse", now.Add(-time.Hour))

	query := BuildListModerationTasksQuery(setup.taskRepo, setup.moderatorRepo, setup.historyRepo)

	// When
	items, err := query()

	// Then: oldest first, with title, status and claimer resolved
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, pendingTask.ID().String(), items[0].ID)
	assert.Equal(t, "Vélo hollandais", items[0].ClassifiedAdTitle)
	assert.Equal(t, pendingTask.CreatedAt(), items[0].CreatedAt)
	assert.Equal(t, TaskStatusPending, items[0].Status)
	assert.Empty(t, items[0].ClaimedBy)

	assert.Equal(t, claimedTask.ID().String(), items[1].ID)
	assert.Equal(t, "Table basse", items[1].ClassifiedAdTitle)
	assert.Equal(t, claimedTask.CreatedAt(), items[1].CreatedAt)
	assert.Equal(t, TaskStatusClaimed, items[1].Status)
	assert.Equal(t, "Jane Doe", items[1].ClaimedBy)
}

func TestListModerationTasksQuery_TitleUsesLastSnapshot(t *testing.T) {
	// Given: an ad re-submitted with a corrected title
	setup := setupQueryTest(t)
	now := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)

	seedTask(t, setup, "ad-1", now)
	history := seedSubmittedHistory(t, setup, "ad-1", "Old title", now.Add(-time.Hour))
	appendEntry(t, history, now, domain.HistoryActionSubmitted, nil, nil, &domain.ClassifiedAdSnapshot{Title: "New title"})
	require.NoError(t, setup.historyRepo.Save(history))

	query := BuildListModerationTasksQuery(setup.taskRepo, setup.moderatorRepo, setup.historyRepo)

	// When
	items, err := query()

	// Then
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "New title", items[0].ClassifiedAdTitle)
}

func TestListModerationTasksQuery_MissingHistoryYieldsEmptyTitle(t *testing.T) {
	// Given: a task whose history has not been fed yet (consumers are async)
	setup := setupQueryTest(t)
	now := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	seedTask(t, setup, "ad-1", now)

	query := BuildListModerationTasksQuery(setup.taskRepo, setup.moderatorRepo, setup.historyRepo)

	// When
	items, err := query()

	// Then
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Empty(t, items[0].ClassifiedAdTitle, "Expected an empty title when no snapshot exists yet")
	assert.Equal(t, TaskStatusPending, items[0].Status)
}

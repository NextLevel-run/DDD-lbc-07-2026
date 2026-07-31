package query

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

func TestGetModerationTaskDetailQuery_FullDetail(t *testing.T) {
	// Given: a claimed task on a re-submitted ad with a full history
	setup := setupQueryTest(t)
	now := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)

	moderator := seedModerator(t, setup, "Jane Doe")
	task := seedClaimedTask(t, setup, "ad-1", moderator, now.Add(-time.Hour), now)

	moderatorID := moderator.ID().String()
	challengeReason := string(domain.ChallengeReasonPriceToVerify)
	history := seedSubmittedHistory(t, setup, "ad-1", "Old title", now.Add(-3*time.Hour))
	appendEntry(t, history, now.Add(-2*time.Hour), domain.HistoryActionChallenged, &moderatorID, &challengeReason, nil)
	appendEntry(t, history, now.Add(-time.Hour), domain.HistoryActionSubmitted, nil, nil, &domain.ClassifiedAdSnapshot{Title: "New title"})
	require.NoError(t, setup.historyRepo.Save(history))

	query := BuildGetModerationTaskDetailQuery(setup.taskRepo, setup.moderatorRepo, setup.historyRepo)

	// When
	view, err := query(task.ID().String())

	// Then: the task itself
	require.NoError(t, err)
	assert.Equal(t, task.ID().String(), view.ID)
	assert.Equal(t, "ad-1", view.ClassifiedAdID)
	assert.Equal(t, task.CreatedAt(), view.CreatedAt)
	assert.Equal(t, TaskStatusClaimed, view.Status)
	assert.Equal(t, "Jane Doe", view.ClaimedBy)
	assert.Equal(t, moderatorID, view.ModeratorID)
	require.NotNil(t, view.ClaimedAt)
	assert.Equal(t, now, *view.ClaimedAt)

	// The full history, in order
	require.Len(t, view.History, 3)
	assert.Equal(t, string(domain.HistoryActionSubmitted), view.History[0].Action)
	require.NotNil(t, view.History[0].Snapshot)
	assert.Equal(t, "Old title", view.History[0].Snapshot.Title)

	assert.Equal(t, string(domain.HistoryActionChallenged), view.History[1].Action)
	require.NotNil(t, view.History[1].ModeratorID)
	assert.Equal(t, moderatorID, *view.History[1].ModeratorID)
	require.NotNil(t, view.History[1].Reason)
	assert.Equal(t, challengeReason, *view.History[1].Reason)
	assert.Nil(t, view.History[1].Snapshot)

	assert.Equal(t, string(domain.HistoryActionSubmitted), view.History[2].Action)
	assert.Equal(t, now.Add(-time.Hour), view.History[2].OccurredAt)

	// The last snapshot reflects the latest submission
	require.NotNil(t, view.LastSnapshot)
	assert.Equal(t, "New title", view.LastSnapshot.Title)
}

func TestGetModerationTaskDetailQuery_PendingTaskWithoutHistory(t *testing.T) {
	// Given: an unclaimed task whose history has not been fed yet
	setup := setupQueryTest(t)
	now := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	task := seedTask(t, setup, "ad-1", now)

	query := BuildGetModerationTaskDetailQuery(setup.taskRepo, setup.moderatorRepo, setup.historyRepo)

	// When
	view, err := query(task.ID().String())

	// Then
	require.NoError(t, err)
	assert.Equal(t, TaskStatusPending, view.Status)
	assert.Empty(t, view.ClaimedBy)
	assert.Empty(t, view.ModeratorID)
	assert.Nil(t, view.ClaimedAt)
	assert.Empty(t, view.History)
	assert.Nil(t, view.LastSnapshot)
}

func TestGetModerationTaskDetailQuery_Errors(t *testing.T) {
	tests := []struct {
		name   string
		taskID string
	}{
		{name: "InvalidTaskIDFormat", taskID: "not-a-uuid"},
		{name: "TaskNotFound", taskID: uuid.New().String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupQueryTest(t)
			query := BuildGetModerationTaskDetailQuery(setup.taskRepo, setup.moderatorRepo, setup.historyRepo)

			// When
			view, err := query(tt.taskID)

			// Then
			require.ErrorIs(t, err, domain.ErrModerationTaskNotFound)
			assert.Equal(t, ModerationTaskDetailView{}, view)
		})
	}
}

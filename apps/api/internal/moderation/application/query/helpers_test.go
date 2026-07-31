package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// queryTestSetup contains the repositories shared by both query use cases.
type queryTestSetup struct {
	taskRepo      *inmemory.InMemoryModerationTaskRepository
	moderatorRepo *inmemory.InMemoryModeratorRepository
	historyRepo   *inmemory.InMemoryClassifiedAdHistoryRepository
}

func setupQueryTest(t *testing.T) *queryTestSetup {
	t.Helper()

	return &queryTestSetup{
		taskRepo:      inmemory.NewInMemoryModerationTaskRepository(),
		moderatorRepo: inmemory.NewInMemoryModeratorRepository(),
		historyRepo:   inmemory.NewInMemoryClassifiedAdHistoryRepository(),
	}
}

// seedModerator creates and stores a moderator with the given full name.
func seedModerator(t *testing.T, setup *queryTestSetup, fullName string) *domain.Moderator {
	t.Helper()

	moderator, err := domain.NewModerator(fullName)
	require.NoError(t, err)
	require.NoError(t, setup.moderatorRepo.Save(moderator))
	return moderator
}

// seedTask creates and stores an unclaimed moderation task for the given ad.
func seedTask(t *testing.T, setup *queryTestSetup, classifiedAdID string, createdAt time.Time) *domain.ModerationTask {
	t.Helper()

	task, err := domain.NewModerationTask(classifiedAdID, createdAt)
	require.NoError(t, err)
	require.NoError(t, setup.taskRepo.Save(task))
	return task
}

// seedClaimedTask creates and stores a task already claimed by the given moderator.
func seedClaimedTask(t *testing.T, setup *queryTestSetup, classifiedAdID string, moderator *domain.Moderator, createdAt, claimedAt time.Time) *domain.ModerationTask {
	t.Helper()

	task := seedTask(t, setup, classifiedAdID, createdAt)
	require.NoError(t, task.Claim(moderator.ID(), claimedAt))
	require.NoError(t, setup.taskRepo.Save(task))
	return task
}

// seedSubmittedHistory creates a history for the ad holding a single submitted
// entry carrying a snapshot with the given title.
func seedSubmittedHistory(t *testing.T, setup *queryTestSetup, classifiedAdID, title string, occurredAt time.Time) *domain.ClassifiedAdHistory {
	t.Helper()

	history, err := domain.NewClassifiedAdHistory(classifiedAdID)
	require.NoError(t, err)
	appendEntry(t, history, occurredAt, domain.HistoryActionSubmitted, nil, nil, &domain.ClassifiedAdSnapshot{
		Title:        title,
		Description:  "Description of " + title,
		PriceInCents: 15000,
		ImageURLs:    []string{"http://img/1.jpg"},
		Category:     "consumer_goods",
		ZipCode:      "75001",
		CityName:     "Paris",
		SellerEmail:  "seller@example.com",
		SellerPseudo: "seller-pseudo",
	})
	require.NoError(t, setup.historyRepo.Save(history))
	return history
}

// appendEntry appends a validated entry to the given history.
func appendEntry(
	t *testing.T,
	history *domain.ClassifiedAdHistory,
	occurredAt time.Time,
	action domain.HistoryAction,
	moderatorID *string,
	reason *string,
	snapshot *domain.ClassifiedAdSnapshot,
) {
	t.Helper()

	entry, err := domain.NewHistoryEntry(occurredAt, action, moderatorID, reason, snapshot)
	require.NoError(t, err)
	history.Append(entry)
}

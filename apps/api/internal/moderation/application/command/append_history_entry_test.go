package command

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/moderation/domain"
)

type appendHistoryTestSetup struct {
	historyRepo *inmemory.InMemoryClassifiedAdHistoryRepository
	command     AppendHistoryEntryCommand
}

func setupAppendHistoryTest(t *testing.T) *appendHistoryTestSetup {
	t.Helper()

	historyRepo := inmemory.NewInMemoryClassifiedAdHistoryRepository()

	return &appendHistoryTestSetup{
		historyRepo: historyRepo,
		command:     BuildAppendHistoryEntryCommand(historyRepo),
	}
}

func TestAppendHistoryEntryCommand_CreatesHistoryOnFirstAppend(t *testing.T) {
	// Given
	setup := setupAppendHistoryTest(t)
	occurredAt := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	snapshot := &domain.ClassifiedAdSnapshot{
		Title:        "Vélo hollandais",
		Description:  "Très bon état, peu servi.",
		PriceInCents: 15000,
		ImageURLs:    []string{"http://img/1.jpg"},
		Category:     "consumer_goods",
		ZipCode:      "75001",
		CityName:     "Paris",
		SellerEmail:  "seller@example.com",
		SellerPseudo: "seller-pseudo",
	}

	// When
	err := setup.command(AppendHistoryEntryCommandArgs{
		ClassifiedAdID: "ad-1",
		OccurredAt:     occurredAt,
		Action:         string(domain.HistoryActionSubmitted),
		Snapshot:       snapshot,
	})

	// Then
	require.NoError(t, err, "Expected the history to be created on first append")

	history, err := setup.historyRepo.FindByClassifiedAdID("ad-1")
	require.NoError(t, err)
	entries := history.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, occurredAt, entries[0].OccurredAt())
	assert.Equal(t, domain.HistoryActionSubmitted, entries[0].Action())
	assert.Nil(t, entries[0].ModeratorID())
	assert.Nil(t, entries[0].Reason())
	require.NotNil(t, entries[0].Snapshot())
	assert.Equal(t, *snapshot, *entries[0].Snapshot())
}

func TestAppendHistoryEntryCommand_AppendsToExistingHistory(t *testing.T) {
	// Given: a history already holding a submitted entry
	setup := setupAppendHistoryTest(t)
	submittedAt := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	require.NoError(t, setup.command(AppendHistoryEntryCommandArgs{
		ClassifiedAdID: "ad-1",
		OccurredAt:     submittedAt,
		Action:         string(domain.HistoryActionSubmitted),
		Snapshot:       &domain.ClassifiedAdSnapshot{Title: "Vélo hollandais"},
	}))

	moderatorID := "moderator-1"
	reason := string(domain.RejectReasonSuspectPrice)
	rejectedAt := submittedAt.Add(time.Hour)

	// When: a moderation action is appended
	err := setup.command(AppendHistoryEntryCommandArgs{
		ClassifiedAdID: "ad-1",
		OccurredAt:     rejectedAt,
		Action:         string(domain.HistoryActionRejected),
		ModeratorID:    &moderatorID,
		Reason:         &reason,
	})

	// Then: entries are kept in order with their respective payloads
	require.NoError(t, err)

	history, err := setup.historyRepo.FindByClassifiedAdID("ad-1")
	require.NoError(t, err)
	entries := history.Entries()
	require.Len(t, entries, 2)

	assert.Equal(t, domain.HistoryActionSubmitted, entries[0].Action())
	assert.Equal(t, domain.HistoryActionRejected, entries[1].Action())
	assert.Equal(t, rejectedAt, entries[1].OccurredAt())
	require.NotNil(t, entries[1].ModeratorID())
	assert.Equal(t, moderatorID, *entries[1].ModeratorID())
	require.NotNil(t, entries[1].Reason())
	assert.Equal(t, reason, *entries[1].Reason())
	assert.Nil(t, entries[1].Snapshot())

	// The derived status and last snapshot follow the log
	status, ok := history.CurrentStatus()
	require.True(t, ok)
	assert.Equal(t, domain.HistoryActionRejected, status)
	require.NotNil(t, history.LastSnapshot())
	assert.Equal(t, "Vélo hollandais", history.LastSnapshot().Title)
}

func TestAppendHistoryEntryCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		args          AppendHistoryEntryCommandArgs
		expectedError error
	}{
		{
			name: "InvalidAction",
			args: AppendHistoryEntryCommandArgs{
				ClassifiedAdID: "ad-1",
				OccurredAt:     time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC),
				Action:         "not-an-action",
			},
			expectedError: domain.ErrInvalidHistoryAction,
		},
		{
			name: "EmptyClassifiedAdID",
			args: AppendHistoryEntryCommandArgs{
				ClassifiedAdID: "",
				OccurredAt:     time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC),
				Action:         string(domain.HistoryActionSubmitted),
			},
			expectedError: domain.ErrEmptyClassifiedAdID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupAppendHistoryTest(t)

			// When
			err := setup.command(tt.args)

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			_, findErr := setup.historyRepo.FindByClassifiedAdID(tt.args.ClassifiedAdID)
			assert.ErrorIs(t, findErr, domain.ErrClassifiedAdHistoryNotFound, "Expected no history to be persisted on error")
		})
	}
}

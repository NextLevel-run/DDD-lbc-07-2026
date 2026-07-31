package domain_test

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newHistoryEntry(t *testing.T, occurredAt time.Time, action domain.HistoryAction, snapshot *domain.ClassifiedAdSnapshot) domain.HistoryEntry {
	t.Helper()
	entry, err := domain.NewHistoryEntry(occurredAt, action, nil, nil, snapshot)
	require.NoError(t, err)
	return entry
}

func TestNewClassifiedAdHistory(t *testing.T) {
	t.Run("valid history starts empty", func(t *testing.T) {
		history, err := domain.NewClassifiedAdHistory("ad-123")
		require.NoError(t, err)
		assert.Equal(t, "ad-123", history.ClassifiedAdID())
		assert.Empty(t, history.Entries())
	})

	t.Run("empty classified ad id is rejected", func(t *testing.T) {
		_, err := domain.NewClassifiedAdHistory("")
		assert.ErrorIs(t, err, domain.ErrEmptyClassifiedAdID)
	})
}

func TestClassifiedAdHistory_Append(t *testing.T) {
	now := time.Now()

	t.Run("entries are kept in append order", func(t *testing.T) {
		history, err := domain.NewClassifiedAdHistory("ad-123")
		require.NoError(t, err)

		history.Append(newHistoryEntry(t, now, domain.HistoryActionSubmitted, newValidSnapshot()))
		history.Append(newHistoryEntry(t, now.Add(time.Hour), domain.HistoryActionApproved, nil))
		history.Append(newHistoryEntry(t, now.Add(2*time.Hour), domain.HistoryActionPublished, nil))

		entries := history.Entries()
		require.Len(t, entries, 3)
		assert.Equal(t, domain.HistoryActionSubmitted, entries[0].Action())
		assert.Equal(t, domain.HistoryActionApproved, entries[1].Action())
		assert.Equal(t, domain.HistoryActionPublished, entries[2].Action())
	})

	t.Run("Entries returns a copy that does not expose the internal log", func(t *testing.T) {
		history, err := domain.NewClassifiedAdHistory("ad-123")
		require.NoError(t, err)
		history.Append(newHistoryEntry(t, now, domain.HistoryActionSubmitted, nil))
		history.Append(newHistoryEntry(t, now.Add(time.Hour), domain.HistoryActionApproved, nil))

		entries := history.Entries()
		entries[0] = newHistoryEntry(t, now, domain.HistoryActionDeleted, nil)

		assert.Equal(t, domain.HistoryActionSubmitted, history.Entries()[0].Action())
	})
}

func TestClassifiedAdHistory_CurrentStatus(t *testing.T) {
	now := time.Now()

	t.Run("empty history has no current status", func(t *testing.T) {
		history, err := domain.NewClassifiedAdHistory("ad-123")
		require.NoError(t, err)

		_, ok := history.CurrentStatus()

		assert.False(t, ok)
	})

	t.Run("current status is derived from the last entry", func(t *testing.T) {
		history, err := domain.NewClassifiedAdHistory("ad-123")
		require.NoError(t, err)
		history.Append(newHistoryEntry(t, now, domain.HistoryActionSubmitted, newValidSnapshot()))
		history.Append(newHistoryEntry(t, now.Add(time.Hour), domain.HistoryActionChallenged, nil))

		status, ok := history.CurrentStatus()

		require.True(t, ok)
		assert.Equal(t, domain.HistoryActionChallenged, status)
	})
}

func TestClassifiedAdHistory_LastSnapshot(t *testing.T) {
	now := time.Now()

	t.Run("no snapshot when history is empty", func(t *testing.T) {
		history, err := domain.NewClassifiedAdHistory("ad-123")
		require.NoError(t, err)

		assert.Nil(t, history.LastSnapshot())
	})

	t.Run("no snapshot when no entry carries one", func(t *testing.T) {
		history, err := domain.NewClassifiedAdHistory("ad-123")
		require.NoError(t, err)
		history.Append(newHistoryEntry(t, now, domain.HistoryActionPublished, nil))

		assert.Nil(t, history.LastSnapshot())
	})

	t.Run("most recent snapshot wins, skipping entries without one", func(t *testing.T) {
		history, err := domain.NewClassifiedAdHistory("ad-123")
		require.NoError(t, err)
		firstSnapshot := newValidSnapshot()
		secondSnapshot := newValidSnapshot()
		secondSnapshot.Title = "An even greater bike"
		history.Append(newHistoryEntry(t, now, domain.HistoryActionSubmitted, firstSnapshot))
		history.Append(newHistoryEntry(t, now.Add(time.Hour), domain.HistoryActionChallenged, nil))
		history.Append(newHistoryEntry(t, now.Add(2*time.Hour), domain.HistoryActionSubmitted, secondSnapshot))
		history.Append(newHistoryEntry(t, now.Add(3*time.Hour), domain.HistoryActionApproved, nil))

		snapshot := history.LastSnapshot()

		require.NotNil(t, snapshot)
		assert.Equal(t, "An even greater bike", snapshot.Title)
	})
}

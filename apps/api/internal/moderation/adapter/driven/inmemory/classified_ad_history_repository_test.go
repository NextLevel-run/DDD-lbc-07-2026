package inmemory

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

func TestInMemoryClassifiedAdHistoryRepository_SaveAndFindByClassifiedAdID(t *testing.T) {
	// Given
	repo := NewInMemoryClassifiedAdHistoryRepository()
	history, err := domain.NewClassifiedAdHistory("ad-1")
	require.NoError(t, err)

	// When
	require.NoError(t, repo.Save(history))

	// Then
	found, err := repo.FindByClassifiedAdID("ad-1")
	require.NoError(t, err)
	assert.Equal(t, history, found)
}

func TestInMemoryClassifiedAdHistoryRepository_SaveUpdatesExistingHistory(t *testing.T) {
	// Given: a saved history that later receives an entry
	repo := NewInMemoryClassifiedAdHistoryRepository()
	history, err := domain.NewClassifiedAdHistory("ad-1")
	require.NoError(t, err)
	require.NoError(t, repo.Save(history))

	entry, err := domain.NewHistoryEntry(
		time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC),
		domain.HistoryActionSubmitted,
		nil,
		nil,
		&domain.ClassifiedAdSnapshot{Title: "Vélo hollandais"},
	)
	require.NoError(t, err)
	history.Append(entry)

	// When
	require.NoError(t, repo.Save(history))

	// Then
	found, err := repo.FindByClassifiedAdID("ad-1")
	require.NoError(t, err)
	require.Len(t, found.Entries(), 1)
	assert.Equal(t, domain.HistoryActionSubmitted, found.Entries()[0].Action())
}

func TestInMemoryClassifiedAdHistoryRepository_FindByClassifiedAdID_NotFound(t *testing.T) {
	// Given
	repo := NewInMemoryClassifiedAdHistoryRepository()

	// When
	_, err := repo.FindByClassifiedAdID("unknown-ad")

	// Then
	assert.ErrorIs(t, err, domain.ErrClassifiedAdHistoryNotFound)
}

func TestInMemoryClassifiedAdHistoryRepository_IsThreadSafe(t *testing.T) {
	// Given
	repo := NewInMemoryClassifiedAdHistoryRepository()

	// When: concurrent writers and readers on distinct ads
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			adID := "ad-" + strconv.Itoa(n)
			history, err := domain.NewClassifiedAdHistory(adID)
			if err != nil {
				return
			}
			_ = repo.Save(history)
			_, _ = repo.FindByClassifiedAdID(adID)
		}(i)
	}
	wg.Wait()

	// Then: every history is retrievable
	for i := 0; i < 20; i++ {
		_, err := repo.FindByClassifiedAdID("ad-" + strconv.Itoa(i))
		assert.NoError(t, err)
	}
}

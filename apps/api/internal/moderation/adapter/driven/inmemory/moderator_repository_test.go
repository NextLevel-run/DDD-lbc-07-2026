package inmemory

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

func TestInMemoryModeratorRepository_SaveAndFindByID(t *testing.T) {
	// Given
	repo := NewInMemoryModeratorRepository()
	moderator, err := domain.NewModerator("Jane Doe")
	require.NoError(t, err)

	// When
	require.NoError(t, repo.Save(moderator))

	// Then
	found, err := repo.FindByID(moderator.ID())
	require.NoError(t, err)
	assert.Equal(t, moderator, found)
}

func TestInMemoryModeratorRepository_FindByID_NotFound(t *testing.T) {
	// Given
	repo := NewInMemoryModeratorRepository()

	// When
	_, err := repo.FindByID(uuid.New())

	// Then
	assert.ErrorIs(t, err, domain.ErrModeratorNotFound)
}

func TestInMemoryModeratorRepository_IsThreadSafe(t *testing.T) {
	// Given
	repo := NewInMemoryModeratorRepository()

	// When: concurrent writers and readers
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			moderator, err := domain.NewModerator("Jane Doe")
			if err != nil {
				return
			}
			_ = repo.Save(moderator)
			_, _ = repo.FindByID(moderator.ID())
		}()
	}
	wg.Wait()
}

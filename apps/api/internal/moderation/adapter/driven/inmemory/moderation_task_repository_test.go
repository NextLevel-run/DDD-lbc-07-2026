package inmemory

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

func newTask(t *testing.T, classifiedAdID string) *domain.ModerationTask {
	t.Helper()

	task, err := domain.NewModerationTask(classifiedAdID, time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	return task
}

func TestInMemoryModerationTaskRepository_SaveAndFindByID(t *testing.T) {
	// Given
	repo := NewInMemoryModerationTaskRepository()
	task := newTask(t, "ad-1")

	// When
	require.NoError(t, repo.Save(task))

	// Then
	found, err := repo.FindByID(task.ID())
	require.NoError(t, err)
	assert.Equal(t, task, found)
}

func TestInMemoryModerationTaskRepository_FindByID_NotFound(t *testing.T) {
	// Given
	repo := NewInMemoryModerationTaskRepository()

	// When
	_, err := repo.FindByID(uuid.New())

	// Then
	assert.ErrorIs(t, err, domain.ErrModerationTaskNotFound)
}

func TestInMemoryModerationTaskRepository_FindAll(t *testing.T) {
	// Given
	repo := NewInMemoryModerationTaskRepository()

	// Empty at first
	tasks, err := repo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// When
	first := newTask(t, "ad-1")
	second := newTask(t, "ad-2")
	require.NoError(t, repo.Save(first))
	require.NoError(t, repo.Save(second))

	// Then
	tasks, err = repo.FindAll()
	require.NoError(t, err)
	assert.ElementsMatch(t, []*domain.ModerationTask{first, second}, tasks)
}

func TestInMemoryModerationTaskRepository_Delete(t *testing.T) {
	// Given
	repo := NewInMemoryModerationTaskRepository()
	task := newTask(t, "ad-1")
	require.NoError(t, repo.Save(task))

	// When
	require.NoError(t, repo.Delete(task.ID()))

	// Then
	_, err := repo.FindByID(task.ID())
	assert.ErrorIs(t, err, domain.ErrModerationTaskNotFound)

	// Deleting again fails: the task is gone
	assert.ErrorIs(t, repo.Delete(task.ID()), domain.ErrModerationTaskNotFound)
}

func TestInMemoryModerationTaskRepository_IsThreadSafe(t *testing.T) {
	// Given
	repo := NewInMemoryModerationTaskRepository()

	// When: concurrent writers and readers
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task := newTask(t, "ad-concurrent")
			_ = repo.Save(task)
			_, _ = repo.FindByID(task.ID())
			_, _ = repo.FindAll()
			_ = repo.Delete(task.ID())
		}()
	}
	wg.Wait()

	// Then: every goroutine deleted its own task
	tasks, err := repo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

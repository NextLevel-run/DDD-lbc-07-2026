package command

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/moderation/domain"
)

type createTaskTestSetup struct {
	taskRepo *inmemory.InMemoryModerationTaskRepository
	clock    *fakeClock
	command  CreateModerationTaskCommand
}

func setupCreateTaskTest(t *testing.T) *createTaskTestSetup {
	t.Helper()

	taskRepo := inmemory.NewInMemoryModerationTaskRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))

	return &createTaskTestSetup{
		taskRepo: taskRepo,
		clock:    clock,
		command:  BuildCreateModerationTaskCommand(taskRepo, clock),
	}
}

func TestCreateModerationTaskCommand_Success(t *testing.T) {
	// Given
	setup := setupCreateTaskTest(t)

	// When
	taskID, err := setup.command(CreateModerationTaskCommandArgs{ClassifiedAdID: "ad-1"})

	// Then
	require.NoError(t, err, "Expected no error when creating a task for a valid ad id")
	require.NotEmpty(t, taskID, "Expected the created task id to be returned")

	// Verify persistence
	id, err := uuid.Parse(taskID)
	require.NoError(t, err)
	stored, err := setup.taskRepo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, "ad-1", stored.ClassifiedAdID())
	assert.Equal(t, setup.clock.Now(), stored.CreatedAt())
	assert.False(t, stored.IsClaimed())
}

func TestCreateModerationTaskCommand_EachSubmissionCreatesANewTask(t *testing.T) {
	// Given
	setup := setupCreateTaskTest(t)

	// When: the same ad is submitted then re-submitted
	firstID, err := setup.command(CreateModerationTaskCommandArgs{ClassifiedAdID: "ad-1"})
	require.NoError(t, err)
	secondID, err := setup.command(CreateModerationTaskCommandArgs{ClassifiedAdID: "ad-1"})
	require.NoError(t, err)

	// Then: two distinct tasks exist
	assert.NotEqual(t, firstID, secondID, "Expected each submission to create a task with a new id")
	tasks, err := setup.taskRepo.FindAll()
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestCreateModerationTaskCommand_EmptyClassifiedAdID(t *testing.T) {
	// Given
	setup := setupCreateTaskTest(t)

	// When
	taskID, err := setup.command(CreateModerationTaskCommandArgs{ClassifiedAdID: ""})

	// Then
	require.ErrorIs(t, err, domain.ErrEmptyClassifiedAdID)
	assert.Empty(t, taskID)

	tasks, err := setup.taskRepo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, tasks, "Expected no task to be persisted on error")
}

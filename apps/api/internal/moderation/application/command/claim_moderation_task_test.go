package command

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"
)

type claimTestSetup struct {
	taskRepo       *inmemory.InMemoryModerationTaskRepository
	moderatorRepo  *inmemory.InMemoryModeratorRepository
	clock          *fakeClock
	eventCollector *eventbustesting.EventCollector
	command        ClaimModerationTaskCommand
}

func setupClaimTest(t *testing.T) *claimTestSetup {
	t.Helper()

	taskRepo := inmemory.NewInMemoryModerationTaskRepository()
	moderatorRepo := inmemory.NewInMemoryModeratorRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("ModerationTaskClaimed", collector.EventHandler())
	require.NoError(t, err)

	return &claimTestSetup{
		taskRepo:       taskRepo,
		moderatorRepo:  moderatorRepo,
		clock:          clock,
		eventCollector: collector,
		command:        BuildClaimModerationTaskCommand(taskRepo, moderatorRepo, clock, bus),
	}
}

func TestClaimModerationTaskCommand_Success(t *testing.T) {
	// Given
	setup := setupClaimTest(t)
	moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
	task := seedTask(t, setup.taskRepo, "ad-1", setup.clock.Now().Add(-time.Hour))

	// When
	err := setup.command(ClaimModerationTaskCommandArgs{
		TaskID:      task.ID().String(),
		ModeratorID: moderator.ID().String(),
	})

	// Then
	require.NoError(t, err, "Expected no error when claiming an unclaimed task")

	// Verify persistence
	stored, err := setup.taskRepo.FindByID(task.ID())
	require.NoError(t, err)
	assert.True(t, stored.IsClaimed())
	require.NotNil(t, stored.ModeratorID())
	assert.Equal(t, moderator.ID(), *stored.ModeratorID())
	require.NotNil(t, stored.ClaimedAt())
	assert.Equal(t, setup.clock.Now(), *stored.ClaimedAt())

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)

	event, ok := events[0].(*domain.ModerationTaskClaimedEvent)
	require.True(t, ok, "Expected event to be *ModerationTaskClaimedEvent")
	assert.Equal(t, task.ID().String(), event.TaskID)
	assert.Equal(t, "ad-1", event.ClassifiedAdID)
	assert.Equal(t, moderator.ID().String(), event.ModeratorID)
	assert.Equal(t, setup.clock.Now(), event.OccurredAt)
}

func TestClaimModerationTaskCommand_AlreadyClaimed(t *testing.T) {
	// Given
	setup := setupClaimTest(t)
	owner := seedModerator(t, setup.moderatorRepo, "Jane Doe")
	other := seedModerator(t, setup.moderatorRepo, "John Smith")
	task := seedClaimedTask(t, setup.taskRepo, "ad-1", owner, setup.clock.Now())
	setup.eventCollector.Clear()

	// When: another moderator tries to claim the same task
	err := setup.command(ClaimModerationTaskCommandArgs{
		TaskID:      task.ID().String(),
		ModeratorID: other.ID().String(),
	})

	// Then
	require.ErrorIs(t, err, domain.ErrTaskAlreadyClaimed)
	assertNoEventsEmitted(t, setup.eventCollector)

	// The claim owner is unchanged
	stored, findErr := setup.taskRepo.FindByID(task.ID())
	require.NoError(t, findErr)
	require.NotNil(t, stored.ModeratorID())
	assert.Equal(t, owner.ID(), *stored.ModeratorID())
}

func TestClaimModerationTaskCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		args          func(t *testing.T, setup *claimTestSetup) ClaimModerationTaskCommandArgs
		expectedError error
	}{
		{
			name: "InvalidTaskIDFormat",
			args: func(t *testing.T, setup *claimTestSetup) ClaimModerationTaskCommandArgs {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				return ClaimModerationTaskCommandArgs{TaskID: "not-a-uuid", ModeratorID: moderator.ID().String()}
			},
			expectedError: domain.ErrModerationTaskNotFound,
		},
		{
			name: "TaskNotFound",
			args: func(t *testing.T, setup *claimTestSetup) ClaimModerationTaskCommandArgs {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				return ClaimModerationTaskCommandArgs{TaskID: uuid.New().String(), ModeratorID: moderator.ID().String()}
			},
			expectedError: domain.ErrModerationTaskNotFound,
		},
		{
			name: "InvalidModeratorIDFormat",
			args: func(t *testing.T, setup *claimTestSetup) ClaimModerationTaskCommandArgs {
				task := seedTask(t, setup.taskRepo, "ad-1", setup.clock.Now())
				return ClaimModerationTaskCommandArgs{TaskID: task.ID().String(), ModeratorID: "not-a-uuid"}
			},
			expectedError: domain.ErrModeratorNotFound,
		},
		{
			name: "ModeratorNotFound",
			args: func(t *testing.T, setup *claimTestSetup) ClaimModerationTaskCommandArgs {
				task := seedTask(t, setup.taskRepo, "ad-1", setup.clock.Now())
				return ClaimModerationTaskCommandArgs{TaskID: task.ID().String(), ModeratorID: uuid.New().String()}
			},
			expectedError: domain.ErrModeratorNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupClaimTest(t)
			args := tt.args(t, setup)

			// When
			err := setup.command(args)

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			assertNoEventsEmitted(t, setup.eventCollector)
		})
	}
}

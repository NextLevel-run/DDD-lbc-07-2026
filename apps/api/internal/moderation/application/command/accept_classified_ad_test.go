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

type acceptTestSetup struct {
	taskRepo       *inmemory.InMemoryModerationTaskRepository
	moderatorRepo  *inmemory.InMemoryModeratorRepository
	clock          *fakeClock
	eventCollector *eventbustesting.EventCollector
	command        AcceptClassifiedAdCommand
}

func setupAcceptTest(t *testing.T) *acceptTestSetup {
	t.Helper()

	taskRepo := inmemory.NewInMemoryModerationTaskRepository()
	moderatorRepo := inmemory.NewInMemoryModeratorRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	require.NoError(t, bus.Subscribe("ModerationTaskCompleted", collector.EventHandler()))
	require.NoError(t, bus.Subscribe("ClassifiedAdApproved", collector.EventHandler()))

	return &acceptTestSetup{
		taskRepo:       taskRepo,
		moderatorRepo:  moderatorRepo,
		clock:          clock,
		eventCollector: collector,
		command:        BuildAcceptClassifiedAdCommand(taskRepo, clock, bus),
	}
}

func TestAcceptClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupAcceptTest(t)
	moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
	task := seedClaimedTask(t, setup.taskRepo, "ad-1", moderator, setup.clock.Now())

	// When
	err := setup.command(AcceptClassifiedAdCommandArgs{
		TaskID:      task.ID().String(),
		ModeratorID: moderator.ID().String(),
	})

	// Then
	require.NoError(t, err, "Expected no error when the claim owner accepts the ad")

	// The task is physically deleted
	_, err = setup.taskRepo.FindByID(task.ID())
	assert.ErrorIs(t, err, domain.ErrModerationTaskNotFound)

	// Verify event emission: completed then approved
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 2)

	completed, ok := events[0].(*domain.ModerationTaskCompletedEvent)
	require.True(t, ok, "Expected first event to be *ModerationTaskCompletedEvent")
	assert.Equal(t, task.ID().String(), completed.TaskID)
	assert.Equal(t, "ad-1", completed.ClassifiedAdID)
	assert.Equal(t, moderator.ID().String(), completed.ModeratorID)
	assert.Equal(t, setup.clock.Now(), completed.OccurredAt)

	approved, ok := events[1].(*domain.ClassifiedAdApprovedEvent)
	require.True(t, ok, "Expected second event to be *ClassifiedAdApprovedEvent")
	assert.Equal(t, "ad-1", approved.ClassifiedAdID)
	assert.Equal(t, moderator.ID().String(), approved.ModeratorID)
	assert.Equal(t, setup.clock.Now(), approved.OccurredAt)
}

func TestAcceptClassifiedAdCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		args          func(t *testing.T, setup *acceptTestSetup) (AcceptClassifiedAdCommandArgs, *domain.ModerationTask)
		expectedError error
	}{
		{
			name: "InvalidTaskIDFormat",
			args: func(t *testing.T, setup *acceptTestSetup) (AcceptClassifiedAdCommandArgs, *domain.ModerationTask) {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				return AcceptClassifiedAdCommandArgs{TaskID: "not-a-uuid", ModeratorID: moderator.ID().String()}, nil
			},
			expectedError: domain.ErrModerationTaskNotFound,
		},
		{
			name: "TaskNotFound",
			args: func(t *testing.T, setup *acceptTestSetup) (AcceptClassifiedAdCommandArgs, *domain.ModerationTask) {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				return AcceptClassifiedAdCommandArgs{TaskID: uuid.New().String(), ModeratorID: moderator.ID().String()}, nil
			},
			expectedError: domain.ErrModerationTaskNotFound,
		},
		{
			name: "InvalidModeratorIDFormat",
			args: func(t *testing.T, setup *acceptTestSetup) (AcceptClassifiedAdCommandArgs, *domain.ModerationTask) {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				task := seedClaimedTask(t, setup.taskRepo, "ad-1", moderator, setup.clock.Now())
				return AcceptClassifiedAdCommandArgs{TaskID: task.ID().String(), ModeratorID: "not-a-uuid"}, task
			},
			expectedError: domain.ErrNotTaskOwner,
		},
		{
			name: "ClaimedByAnotherModerator",
			args: func(t *testing.T, setup *acceptTestSetup) (AcceptClassifiedAdCommandArgs, *domain.ModerationTask) {
				owner := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				other := seedModerator(t, setup.moderatorRepo, "John Smith")
				task := seedClaimedTask(t, setup.taskRepo, "ad-1", owner, setup.clock.Now())
				return AcceptClassifiedAdCommandArgs{TaskID: task.ID().String(), ModeratorID: other.ID().String()}, task
			},
			expectedError: domain.ErrNotTaskOwner,
		},
		{
			name: "UnclaimedTask",
			args: func(t *testing.T, setup *acceptTestSetup) (AcceptClassifiedAdCommandArgs, *domain.ModerationTask) {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				task := seedTask(t, setup.taskRepo, "ad-1", setup.clock.Now())
				return AcceptClassifiedAdCommandArgs{TaskID: task.ID().String(), ModeratorID: moderator.ID().String()}, task
			},
			expectedError: domain.ErrNotTaskOwner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupAcceptTest(t)
			args, task := tt.args(t, setup)

			// When
			err := setup.command(args)

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			assertNoEventsEmitted(t, setup.eventCollector)
			if task != nil {
				assertTaskStillStored(t, setup.taskRepo, task)
			}
		})
	}
}

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

type challengeTestSetup struct {
	taskRepo       *inmemory.InMemoryModerationTaskRepository
	moderatorRepo  *inmemory.InMemoryModeratorRepository
	clock          *fakeClock
	eventCollector *eventbustesting.EventCollector
	command        ChallengeClassifiedAdCommand
}

func setupChallengeTest(t *testing.T) *challengeTestSetup {
	t.Helper()

	taskRepo := inmemory.NewInMemoryModerationTaskRepository()
	moderatorRepo := inmemory.NewInMemoryModeratorRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	require.NoError(t, bus.Subscribe("ModerationTaskCompleted", collector.EventHandler()))
	require.NoError(t, bus.Subscribe("ClassifiedAdChallenged", collector.EventHandler()))

	return &challengeTestSetup{
		taskRepo:       taskRepo,
		moderatorRepo:  moderatorRepo,
		clock:          clock,
		eventCollector: collector,
		command:        BuildChallengeClassifiedAdCommand(taskRepo, clock, bus),
	}
}

func TestChallengeClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupChallengeTest(t)
	moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
	task := seedClaimedTask(t, setup.taskRepo, "ad-1", moderator, setup.clock.Now())

	// When
	err := setup.command(ChallengeClassifiedAdCommandArgs{
		TaskID:      task.ID().String(),
		ModeratorID: moderator.ID().String(),
		Reason:      string(domain.ChallengeReasonCategoryToFix),
	})

	// Then
	require.NoError(t, err, "Expected no error when the claim owner challenges the ad")

	// The task is physically deleted
	_, err = setup.taskRepo.FindByID(task.ID())
	assert.ErrorIs(t, err, domain.ErrModerationTaskNotFound)

	// Verify event emission: completed then challenged
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 2)

	completed, ok := events[0].(*domain.ModerationTaskCompletedEvent)
	require.True(t, ok, "Expected first event to be *ModerationTaskCompletedEvent")
	assert.Equal(t, task.ID().String(), completed.TaskID)
	assert.Equal(t, "ad-1", completed.ClassifiedAdID)
	assert.Equal(t, moderator.ID().String(), completed.ModeratorID)
	assert.Equal(t, setup.clock.Now(), completed.OccurredAt)

	challenged, ok := events[1].(*domain.ClassifiedAdChallengedEvent)
	require.True(t, ok, "Expected second event to be *ClassifiedAdChallengedEvent")
	assert.Equal(t, "ad-1", challenged.ClassifiedAdID)
	assert.Equal(t, moderator.ID().String(), challenged.ModeratorID)
	assert.Equal(t, string(domain.ChallengeReasonCategoryToFix), challenged.Reason)
	assert.Equal(t, setup.clock.Now(), challenged.OccurredAt)
}

func TestChallengeClassifiedAdCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		args          func(t *testing.T, setup *challengeTestSetup) (ChallengeClassifiedAdCommandArgs, *domain.ModerationTask)
		expectedError error
	}{
		{
			name: "InvalidTaskIDFormat",
			args: func(t *testing.T, setup *challengeTestSetup) (ChallengeClassifiedAdCommandArgs, *domain.ModerationTask) {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				return ChallengeClassifiedAdCommandArgs{
					TaskID:      "not-a-uuid",
					ModeratorID: moderator.ID().String(),
					Reason:      string(domain.ChallengeReasonPriceToVerify),
				}, nil
			},
			expectedError: domain.ErrModerationTaskNotFound,
		},
		{
			name: "TaskNotFound",
			args: func(t *testing.T, setup *challengeTestSetup) (ChallengeClassifiedAdCommandArgs, *domain.ModerationTask) {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				return ChallengeClassifiedAdCommandArgs{
					TaskID:      uuid.New().String(),
					ModeratorID: moderator.ID().String(),
					Reason:      string(domain.ChallengeReasonPriceToVerify),
				}, nil
			},
			expectedError: domain.ErrModerationTaskNotFound,
		},
		{
			name: "InvalidReason",
			args: func(t *testing.T, setup *challengeTestSetup) (ChallengeClassifiedAdCommandArgs, *domain.ModerationTask) {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				task := seedClaimedTask(t, setup.taskRepo, "ad-1", moderator, setup.clock.Now())
				return ChallengeClassifiedAdCommandArgs{
					TaskID:      task.ID().String(),
					ModeratorID: moderator.ID().String(),
					Reason:      "not-a-reason",
				}, task
			},
			expectedError: domain.ErrInvalidChallengeReason,
		},
		{
			name: "RejectReasonIsNotAChallengeReason",
			args: func(t *testing.T, setup *challengeTestSetup) (ChallengeClassifiedAdCommandArgs, *domain.ModerationTask) {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				task := seedClaimedTask(t, setup.taskRepo, "ad-1", moderator, setup.clock.Now())
				return ChallengeClassifiedAdCommandArgs{
					TaskID:      task.ID().String(),
					ModeratorID: moderator.ID().String(),
					Reason:      string(domain.RejectReasonSuspectPrice),
				}, task
			},
			expectedError: domain.ErrInvalidChallengeReason,
		},
		{
			name: "ClaimedByAnotherModerator",
			args: func(t *testing.T, setup *challengeTestSetup) (ChallengeClassifiedAdCommandArgs, *domain.ModerationTask) {
				owner := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				other := seedModerator(t, setup.moderatorRepo, "John Smith")
				task := seedClaimedTask(t, setup.taskRepo, "ad-1", owner, setup.clock.Now())
				return ChallengeClassifiedAdCommandArgs{
					TaskID:      task.ID().String(),
					ModeratorID: other.ID().String(),
					Reason:      string(domain.ChallengeReasonPriceToVerify),
				}, task
			},
			expectedError: domain.ErrNotTaskOwner,
		},
		{
			name: "UnclaimedTask",
			args: func(t *testing.T, setup *challengeTestSetup) (ChallengeClassifiedAdCommandArgs, *domain.ModerationTask) {
				moderator := seedModerator(t, setup.moderatorRepo, "Jane Doe")
				task := seedTask(t, setup.taskRepo, "ad-1", setup.clock.Now())
				return ChallengeClassifiedAdCommandArgs{
					TaskID:      task.ID().String(),
					ModeratorID: moderator.ID().String(),
					Reason:      string(domain.ChallengeReasonPriceToVerify),
				}, task
			},
			expectedError: domain.ErrNotTaskOwner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupChallengeTest(t)
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

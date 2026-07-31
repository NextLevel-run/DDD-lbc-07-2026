package command

import (
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// ClaimModerationTaskCommandArgs contains input data for the command.
type ClaimModerationTaskCommandArgs struct {
	TaskID      string
	ModeratorID string
}

// ClaimModerationTaskCommand is the command function type.
type ClaimModerationTaskCommand func(args ClaimModerationTaskCommandArgs) error

// BuildClaimModerationTaskCommand builds a command with dependencies injected.
// It locks an unclaimed task exclusively for a moderator and emits a
// ModerationTaskClaimedEvent.
func BuildClaimModerationTaskCommand(
	taskRepo domain.ModerationTaskRepository,
	moderatorRepo domain.ModeratorRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) ClaimModerationTaskCommand {
	return func(args ClaimModerationTaskCommandArgs) error {
		taskID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return domain.ErrModerationTaskNotFound
		}

		task, err := taskRepo.FindByID(taskID)
		if err != nil {
			return err
		}

		moderatorID, err := uuid.Parse(args.ModeratorID)
		if err != nil {
			return domain.ErrModeratorNotFound
		}

		// The moderator must be known: its ID is recorded on the task and its
		// full name is resolved by the list query for claimed tasks.
		if _, err := moderatorRepo.FindByID(moderatorID); err != nil {
			return err
		}

		if err := task.Claim(moderatorID, clock.Now()); err != nil {
			return err
		}

		if err := taskRepo.Save(task); err != nil {
			return err
		}

		event := domain.NewModerationTaskClaimedEventFromTask(task)
		return eventBus.Publish(event)
	}
}

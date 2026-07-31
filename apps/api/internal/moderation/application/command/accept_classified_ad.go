package command

import (
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// AcceptClassifiedAdCommandArgs contains input data for the command.
type AcceptClassifiedAdCommandArgs struct {
	TaskID      string
	ModeratorID string
}

// AcceptClassifiedAdCommand is the command function type.
type AcceptClassifiedAdCommand func(args AcceptClassifiedAdCommandArgs) error

// BuildAcceptClassifiedAdCommand builds a command with dependencies injected.
// The moderator holding the claim approves the classified ad: the task is
// physically deleted (the audit trail lives in ClassifiedAdHistory) and a
// ModerationTaskCompletedEvent plus an internal ClassifiedAdApprovedEvent are
// emitted — the latter is relayed to the public bus by the publisher.
func BuildAcceptClassifiedAdCommand(
	taskRepo domain.ModerationTaskRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) AcceptClassifiedAdCommand {
	return func(args AcceptClassifiedAdCommandArgs) error {
		taskID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return domain.ErrModerationTaskNotFound
		}

		task, err := taskRepo.FindByID(taskID)
		if err != nil {
			return err
		}

		// A malformed moderator ID can never match the claim owner.
		moderatorID, err := uuid.Parse(args.ModeratorID)
		if err != nil {
			return domain.ErrNotTaskOwner
		}

		if err := task.Complete(moderatorID); err != nil {
			return err
		}

		// Build the events before deleting: the builders read from the task.
		now := clock.Now()
		completedEvent := domain.NewModerationTaskCompletedEventFromTask(task, now)
		approvedEvent := domain.NewClassifiedAdApprovedEventFromTask(task, now)

		if err := taskRepo.Delete(task.ID()); err != nil {
			return err
		}

		if err := eventBus.Publish(completedEvent); err != nil {
			return err
		}
		return eventBus.Publish(approvedEvent)
	}
}

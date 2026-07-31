package command

import (
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// RejectClassifiedAdCommandArgs contains input data for the command.
type RejectClassifiedAdCommandArgs struct {
	TaskID      string
	ModeratorID string
	Reason      string
}

// RejectClassifiedAdCommand is the command function type.
type RejectClassifiedAdCommand func(args RejectClassifiedAdCommandArgs) error

// BuildRejectClassifiedAdCommand builds a command with dependencies injected.
// The moderator holding the claim rejects the classified ad with a predefined
// reason: the task is physically deleted and a ModerationTaskCompletedEvent
// plus an internal ClassifiedAdRejectedEvent are emitted — the latter is
// relayed to the public bus by the publisher.
func BuildRejectClassifiedAdCommand(
	taskRepo domain.ModerationTaskRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) RejectClassifiedAdCommand {
	return func(args RejectClassifiedAdCommandArgs) error {
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

		reason, err := domain.NewRejectReason(args.Reason)
		if err != nil {
			return err
		}

		if err := task.Complete(moderatorID); err != nil {
			return err
		}

		// Build the events before deleting: the builders read from the task.
		now := clock.Now()
		completedEvent := domain.NewModerationTaskCompletedEventFromTask(task, now)
		rejectedEvent := domain.NewClassifiedAdRejectedEventFromTask(task, reason, now)

		if err := taskRepo.Delete(task.ID()); err != nil {
			return err
		}

		if err := eventBus.Publish(completedEvent); err != nil {
			return err
		}
		return eventBus.Publish(rejectedEvent)
	}
}

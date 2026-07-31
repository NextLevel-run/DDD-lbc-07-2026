package command

import (
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"

	"github.com/google/uuid"
)

// ChallengeClassifiedAdCommandArgs contains input data for the command.
type ChallengeClassifiedAdCommandArgs struct {
	TaskID      string
	ModeratorID string
	Reason      string
}

// ChallengeClassifiedAdCommand is the command function type.
type ChallengeClassifiedAdCommand func(args ChallengeClassifiedAdCommandArgs) error

// BuildChallengeClassifiedAdCommand builds a command with dependencies injected.
// The moderator holding the claim asks the seller for corrections with a
// predefined reason: the task is physically deleted and a
// ModerationTaskCompletedEvent plus an internal ClassifiedAdChallengedEvent are
// emitted — the latter is relayed to the public bus by the publisher.
func BuildChallengeClassifiedAdCommand(
	taskRepo domain.ModerationTaskRepository,
	clock domain.Clock,
	eventBus eventbus.Bus,
) ChallengeClassifiedAdCommand {
	return func(args ChallengeClassifiedAdCommandArgs) error {
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

		reason, err := domain.NewChallengeReason(args.Reason)
		if err != nil {
			return err
		}

		if err := task.Complete(moderatorID); err != nil {
			return err
		}

		// Build the events before deleting: the builders read from the task.
		now := clock.Now()
		completedEvent := domain.NewModerationTaskCompletedEventFromTask(task, now)
		challengedEvent := domain.NewClassifiedAdChallengedEventFromTask(task, reason, now)

		if err := taskRepo.Delete(task.ID()); err != nil {
			return err
		}

		if err := eventBus.Publish(completedEvent); err != nil {
			return err
		}
		return eventBus.Publish(challengedEvent)
	}
}

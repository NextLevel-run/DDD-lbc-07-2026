package command

import (
	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// CreateModerationTaskCommandArgs contains input data for the command.
type CreateModerationTaskCommandArgs struct {
	ClassifiedAdID string
}

// CreateModerationTaskCommand is the command function type. It returns the ID
// of the created task.
type CreateModerationTaskCommand func(args CreateModerationTaskCommandArgs) (string, error)

// BuildCreateModerationTaskCommand builds a command with dependencies injected.
// It is an internal use case driven by the consumers of the public
// ClassifiedAdSubmitted and ClassifiedAdEdited events: each submission or
// re-submission of an ad enqueues a brand new task (new ID).
func BuildCreateModerationTaskCommand(
	taskRepo domain.ModerationTaskRepository,
	clock domain.Clock,
) CreateModerationTaskCommand {
	return func(args CreateModerationTaskCommandArgs) (string, error) {
		task, err := domain.NewModerationTask(args.ClassifiedAdID, clock.Now())
		if err != nil {
			return "", err
		}

		if err := taskRepo.Save(task); err != nil {
			return "", err
		}

		return task.ID().String(), nil
	}
}

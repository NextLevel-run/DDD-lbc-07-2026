package command

import (
	"errors"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// AppendHistoryEntryCommandArgs contains input data for the command.
// ModeratorID, Reason and Snapshot are optional and depend on the action being
// recorded (moderation actions carry a moderator, reject/challenge a reason,
// submitted/edited a snapshot).
type AppendHistoryEntryCommandArgs struct {
	ClassifiedAdID string
	OccurredAt     time.Time
	Action         string
	ModeratorID    *string
	Reason         *string
	Snapshot       *domain.ClassifiedAdSnapshot
}

// AppendHistoryEntryCommand is the command function type.
type AppendHistoryEntryCommand func(args AppendHistoryEntryCommandArgs) error

// BuildAppendHistoryEntryCommand builds a command with dependencies injected.
// It is an internal use case driven by the consumers of the public ClassifiedAd
// and Moderation events: it appends an entry to the ad's append-only history,
// creating the history on first append.
func BuildAppendHistoryEntryCommand(
	historyRepo domain.ClassifiedAdHistoryRepository,
) AppendHistoryEntryCommand {
	return func(args AppendHistoryEntryCommandArgs) error {
		action, err := domain.NewHistoryAction(args.Action)
		if err != nil {
			return err
		}

		entry, err := domain.NewHistoryEntry(args.OccurredAt, action, args.ModeratorID, args.Reason, args.Snapshot)
		if err != nil {
			return err
		}

		history, err := historyRepo.FindByClassifiedAdID(args.ClassifiedAdID)
		if errors.Is(err, domain.ErrClassifiedAdHistoryNotFound) {
			history, err = domain.NewClassifiedAdHistory(args.ClassifiedAdID)
		}
		if err != nil {
			return err
		}

		history.Append(entry)

		return historyRepo.Save(history)
	}
}

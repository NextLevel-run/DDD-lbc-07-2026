package query

import (
	"errors"
	"sort"
	"time"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// Task statuses exposed by the moderation read models.
const (
	TaskStatusPending = "pending"
	TaskStatusClaimed = "claimed"
)

// ModerationTaskListItem is the read model for one row of the moderation queue.
type ModerationTaskListItem struct {
	ID                string
	ClassifiedAdTitle string
	CreatedAt         time.Time
	Status            string // "pending" or "claimed"
	ClaimedBy         string // moderator full name, empty while pending
}

// ListModerationTasksQuery lists every active moderation task (pending and claimed).
type ListModerationTasksQuery func() ([]ModerationTaskListItem, error)

// BuildListModerationTasksQuery wires the ListModerationTasksQuery use case.
// The ad title comes from the last snapshot in the ad's history; the claimer
// name is resolved through the moderator repository.
func BuildListModerationTasksQuery(
	taskRepo domain.ModerationTaskRepository,
	moderatorRepo domain.ModeratorRepository,
	historyRepo domain.ClassifiedAdHistoryRepository,
) ListModerationTasksQuery {
	return func() ([]ModerationTaskListItem, error) {
		tasks, err := taskRepo.FindAll()
		if err != nil {
			return nil, err
		}

		items := make([]ModerationTaskListItem, 0, len(tasks))
		for _, task := range tasks {
			title, err := resolveClassifiedAdTitle(historyRepo, task.ClassifiedAdID())
			if err != nil {
				return nil, err
			}

			status := TaskStatusPending
			claimedBy := ""
			if task.IsClaimed() {
				status = TaskStatusClaimed
				claimedBy, err = resolveModeratorFullName(moderatorRepo, task)
				if err != nil {
					return nil, err
				}
			}

			items = append(items, ModerationTaskListItem{
				ID:                task.ID().String(),
				ClassifiedAdTitle: title,
				CreatedAt:         task.CreatedAt(),
				Status:            status,
				ClaimedBy:         claimedBy,
			})
		}

		// Oldest first: moderators process the queue in arrival order. The ID
		// tie-break keeps the output deterministic.
		sort.Slice(items, func(i, j int) bool {
			if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
				return items[i].CreatedAt.Before(items[j].CreatedAt)
			}
			return items[i].ID < items[j].ID
		})

		return items, nil
	}
}

// resolveClassifiedAdTitle returns the ad title from the last snapshot of its
// history, or an empty string when no history or snapshot exists yet (the
// history is fed asynchronously by event consumers).
func resolveClassifiedAdTitle(historyRepo domain.ClassifiedAdHistoryRepository, classifiedAdID string) (string, error) {
	history, err := historyRepo.FindByClassifiedAdID(classifiedAdID)
	if errors.Is(err, domain.ErrClassifiedAdHistoryNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	snapshot := history.LastSnapshot()
	if snapshot == nil {
		return "", nil
	}
	return snapshot.Title, nil
}

// resolveModeratorFullName returns the full name of the moderator holding the
// claim on the given task.
func resolveModeratorFullName(moderatorRepo domain.ModeratorRepository, task *domain.ModerationTask) (string, error) {
	moderator, err := moderatorRepo.FindByID(*task.ModeratorID())
	if err != nil {
		return "", err
	}
	return moderator.FullName(), nil
}

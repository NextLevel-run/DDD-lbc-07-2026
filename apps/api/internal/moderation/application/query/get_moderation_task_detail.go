package query

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// ClassifiedAdSnapshotView is the read model for the ad content captured at
// submission or edition time.
type ClassifiedAdSnapshotView struct {
	Title        string
	Description  string
	PriceInCents int64
	ImageURLs    []string
	Category     string
	ZipCode      string
	CityName     string
	SellerEmail  string
	SellerPseudo string
}

// HistoryEntryView is the read model for one entry of a ClassifiedAdHistory.
type HistoryEntryView struct {
	OccurredAt  time.Time
	Action      string
	ModeratorID *string
	Reason      *string
	Snapshot    *ClassifiedAdSnapshotView
}

// ModerationTaskDetailView is the read model for a single moderation task:
// the task itself, the full history of the ad and its last content snapshot.
type ModerationTaskDetailView struct {
	ID             string
	ClassifiedAdID string
	CreatedAt      time.Time
	Status         string // "pending" or "claimed"
	ClaimedBy      string // moderator full name, empty while pending
	ModeratorID    string // empty while pending
	ClaimedAt      *time.Time
	History        []HistoryEntryView
	LastSnapshot   *ClassifiedAdSnapshotView
}

// GetModerationTaskDetailQuery retrieves the detail view of a moderation task by id.
type GetModerationTaskDetailQuery func(taskID string) (ModerationTaskDetailView, error)

// BuildGetModerationTaskDetailQuery wires the GetModerationTaskDetailQuery use
// case. Returns ErrModerationTaskNotFound when the task does not exist; a
// missing history simply yields empty entries (it is fed asynchronously by
// event consumers).
func BuildGetModerationTaskDetailQuery(
	taskRepo domain.ModerationTaskRepository,
	moderatorRepo domain.ModeratorRepository,
	historyRepo domain.ClassifiedAdHistoryRepository,
) GetModerationTaskDetailQuery {
	return func(taskID string) (ModerationTaskDetailView, error) {
		id, err := uuid.Parse(taskID)
		if err != nil {
			return ModerationTaskDetailView{}, domain.ErrModerationTaskNotFound
		}

		task, err := taskRepo.FindByID(id)
		if err != nil {
			return ModerationTaskDetailView{}, err
		}

		view := ModerationTaskDetailView{
			ID:             task.ID().String(),
			ClassifiedAdID: task.ClassifiedAdID(),
			CreatedAt:      task.CreatedAt(),
			Status:         TaskStatusPending,
			ClaimedAt:      task.ClaimedAt(),
			History:        []HistoryEntryView{},
		}

		if task.IsClaimed() {
			view.Status = TaskStatusClaimed
			view.ModeratorID = task.ModeratorID().String()
			view.ClaimedBy, err = resolveModeratorFullName(moderatorRepo, task)
			if err != nil {
				return ModerationTaskDetailView{}, err
			}
		}

		history, err := historyRepo.FindByClassifiedAdID(task.ClassifiedAdID())
		if errors.Is(err, domain.ErrClassifiedAdHistoryNotFound) {
			return view, nil
		}
		if err != nil {
			return ModerationTaskDetailView{}, err
		}

		for _, entry := range history.Entries() {
			view.History = append(view.History, HistoryEntryView{
				OccurredAt:  entry.OccurredAt(),
				Action:      string(entry.Action()),
				ModeratorID: entry.ModeratorID(),
				Reason:      entry.Reason(),
				Snapshot:    newClassifiedAdSnapshotView(entry.Snapshot()),
			})
		}
		view.LastSnapshot = newClassifiedAdSnapshotView(history.LastSnapshot())

		return view, nil
	}
}

// newClassifiedAdSnapshotView maps a domain snapshot to its read model, or nil.
func newClassifiedAdSnapshotView(snapshot *domain.ClassifiedAdSnapshot) *ClassifiedAdSnapshotView {
	if snapshot == nil {
		return nil
	}
	return &ClassifiedAdSnapshotView{
		Title:        snapshot.Title,
		Description:  snapshot.Description,
		PriceInCents: snapshot.PriceInCents,
		ImageURLs:    snapshot.ImageURLs,
		Category:     snapshot.Category,
		ZipCode:      snapshot.ZipCode,
		CityName:     snapshot.CityName,
		SellerEmail:  snapshot.SellerEmail,
		SellerPseudo: snapshot.SellerPseudo,
	}
}

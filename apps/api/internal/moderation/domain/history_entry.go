package domain

import "time"

// HistoryAction represents the kind of event recorded in a ClassifiedAdHistory entry.
type HistoryAction string

const (
	HistoryActionSubmitted  HistoryAction = "submitted"  // ad submitted or re-submitted
	HistoryActionApproved   HistoryAction = "approved"   // approved by a moderator
	HistoryActionRejected   HistoryAction = "rejected"   // rejected by a moderator
	HistoryActionChallenged HistoryAction = "challenged" // challenged by a moderator
	HistoryActionPublished  HistoryAction = "published"  // ad published
	HistoryActionDeleted    HistoryAction = "deleted"    // ad deleted
	HistoryActionExpired    HistoryAction = "expired"    // ad expired
)

// NewHistoryAction validates and builds a HistoryAction from a raw string.
func NewHistoryAction(s string) (HistoryAction, error) {
	switch HistoryAction(s) {
	case HistoryActionSubmitted, HistoryActionApproved, HistoryActionRejected,
		HistoryActionChallenged, HistoryActionPublished, HistoryActionDeleted,
		HistoryActionExpired:
		return HistoryAction(s), nil
	default:
		return "", ErrInvalidHistoryAction
	}
}

// HistoryEntry is a single immutable record in a ClassifiedAdHistory log.
// moderatorID is set for moderation actions, reason for reject/challenge,
// and snapshot for submitted/edited entries.
type HistoryEntry struct {
	occurredAt  time.Time
	action      HistoryAction
	moderatorID *string
	reason      *string
	snapshot    *ClassifiedAdSnapshot
}

// NewHistoryEntry validates and builds a HistoryEntry. moderatorID, reason and
// snapshot are optional and depend on the action being recorded.
func NewHistoryEntry(
	occurredAt time.Time,
	action HistoryAction,
	moderatorID *string,
	reason *string,
	snapshot *ClassifiedAdSnapshot,
) (HistoryEntry, error) {
	if _, err := NewHistoryAction(string(action)); err != nil {
		return HistoryEntry{}, err
	}
	return HistoryEntry{
		occurredAt:  occurredAt,
		action:      action,
		moderatorID: moderatorID,
		reason:      reason,
		snapshot:    snapshot,
	}, nil
}

// OccurredAt returns the timestamp of the recorded event.
func (e HistoryEntry) OccurredAt() time.Time {
	return e.occurredAt
}

// Action returns the kind of event recorded.
func (e HistoryEntry) Action() HistoryAction {
	return e.action
}

// ModeratorID returns the moderator involved, if the entry records a moderation action.
func (e HistoryEntry) ModeratorID() *string {
	return e.moderatorID
}

// Reason returns the reject/challenge reason, if applicable.
func (e HistoryEntry) Reason() *string {
	return e.reason
}

// Snapshot returns the ad content captured at submission/edition, if applicable.
func (e HistoryEntry) Snapshot() *ClassifiedAdSnapshot {
	return e.snapshot
}

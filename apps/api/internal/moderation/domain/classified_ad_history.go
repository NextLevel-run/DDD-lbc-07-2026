package domain

// ClassifiedAdHistory is the aggregate root recording the full moderation
// history of a classified ad as an append-only event log. The ad's current
// status (as known by this bounded context) is derived from the last entry,
// never stored.
type ClassifiedAdHistory struct {
	classifiedAdID string
	entries        []HistoryEntry
}

// NewClassifiedAdHistory validates and builds an empty history for a classified ad.
func NewClassifiedAdHistory(classifiedAdID string) (*ClassifiedAdHistory, error) {
	if classifiedAdID == "" {
		return nil, ErrEmptyClassifiedAdID
	}
	return &ClassifiedAdHistory{classifiedAdID: classifiedAdID}, nil
}

// ClassifiedAdID returns the ID of the classified ad this history belongs to.
func (h *ClassifiedAdHistory) ClassifiedAdID() string {
	return h.classifiedAdID
}

// Entries returns a copy of the chronological entry list, preserving the
// append-only invariant (callers cannot mutate the log).
func (h *ClassifiedAdHistory) Entries() []HistoryEntry {
	entries := make([]HistoryEntry, len(h.entries))
	copy(entries, h.entries)
	return entries
}

// Append records a new entry at the end of the log. Entries are never updated
// nor removed.
func (h *ClassifiedAdHistory) Append(entry HistoryEntry) {
	h.entries = append(h.entries, entry)
}

// CurrentStatus derives the ad's current status from the last entry of the
// log. ok is false when the history is empty.
func (h *ClassifiedAdHistory) CurrentStatus() (action HistoryAction, ok bool) {
	if len(h.entries) == 0 {
		return "", false
	}
	return h.entries[len(h.entries)-1].Action(), true
}

// LastSnapshot returns the most recent ad content snapshot in the log
// (submitted/edited entries carry one), or nil if no entry has a snapshot.
func (h *ClassifiedAdHistory) LastSnapshot() *ClassifiedAdSnapshot {
	for i := len(h.entries) - 1; i >= 0; i-- {
		if snapshot := h.entries[i].Snapshot(); snapshot != nil {
			return snapshot
		}
	}
	return nil
}

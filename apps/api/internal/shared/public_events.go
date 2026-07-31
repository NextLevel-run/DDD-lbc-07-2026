// Package shared holds the public integration contracts exchanged between
// bounded contexts: the public event DTOs and the public event bus abstraction.
// These types are deliberately decoupled from each bounded context's internal
// domain events — they are the only coupling point between contexts.
package shared

import "time"

// Public event type names. Publishers and consumers must use these constants
// instead of hardcoding event type strings.
const (
	ClassifiedAdSubmittedEventType  = "ClassifiedAdSubmitted"
	ClassifiedAdEditedEventType     = "ClassifiedAdEdited"
	ClassifiedAdPublishedEventType  = "ClassifiedAdPublished"
	ClassifiedAdDeletedEventType    = "ClassifiedAdDeleted"
	ClassifiedAdExpiredEventType    = "ClassifiedAdExpired"
	ClassifiedAdApprovedEventType   = "ClassifiedAdApproved"
	ClassifiedAdRejectedEventType   = "ClassifiedAdRejected"
	ClassifiedAdChallengedEventType = "ClassifiedAdChallenged"
)

// --- ClassifiedAd → Public ---

// ClassifiedAdSubmitted is published when a new classified ad is submitted
// and awaits moderation.
type ClassifiedAdSubmitted struct {
	ClassifiedAdID string
	Title          string
	Description    string
	PriceInCents   int64
	ImageURLs      []string
	Category       string
	ZipCode        string
	CityName       string
	SellerEmail    string
	SellerPseudo   string
	OccurredAt     time.Time
}

// EventType returns the public event type name.
func (e *ClassifiedAdSubmitted) EventType() string {
	return ClassifiedAdSubmittedEventType
}

// ClassifiedAdEdited is published when a seller edits a challenged classified
// ad, re-submitting it for moderation.
type ClassifiedAdEdited struct {
	ClassifiedAdID string
	Title          string
	Description    string
	PriceInCents   int64
	ImageURLs      []string
	Category       string
	ZipCode        string
	CityName       string
	SellerEmail    string
	SellerPseudo   string
	OccurredAt     time.Time
}

// EventType returns the public event type name.
func (e *ClassifiedAdEdited) EventType() string {
	return ClassifiedAdEditedEventType
}

// ClassifiedAdPublished is published when an approved classified ad goes online.
type ClassifiedAdPublished struct {
	ClassifiedAdID string
	OccurredAt     time.Time
}

// EventType returns the public event type name.
func (e *ClassifiedAdPublished) EventType() string {
	return ClassifiedAdPublishedEventType
}

// ClassifiedAdDeleted is published when a classified ad is deleted.
type ClassifiedAdDeleted struct {
	ClassifiedAdID string
	Reason         string
	OccurredAt     time.Time
}

// EventType returns the public event type name.
func (e *ClassifiedAdDeleted) EventType() string {
	return ClassifiedAdDeletedEventType
}

// ClassifiedAdExpired is published when a classified ad expires after its
// lifetime elapsed.
type ClassifiedAdExpired struct {
	ClassifiedAdID string
	OccurredAt     time.Time
}

// EventType returns the public event type name.
func (e *ClassifiedAdExpired) EventType() string {
	return ClassifiedAdExpiredEventType
}

// --- Moderation → Public ---

// ClassifiedAdApproved is published when a moderator approves a classified ad.
type ClassifiedAdApproved struct {
	ClassifiedAdID string
	ModeratorID    string
	OccurredAt     time.Time
}

// EventType returns the public event type name.
func (e *ClassifiedAdApproved) EventType() string {
	return ClassifiedAdApprovedEventType
}

// ClassifiedAdRejected is published when a moderator rejects a classified ad.
type ClassifiedAdRejected struct {
	ClassifiedAdID string
	ModeratorID    string
	Reason         string
	OccurredAt     time.Time
}

// EventType returns the public event type name.
func (e *ClassifiedAdRejected) EventType() string {
	return ClassifiedAdRejectedEventType
}

// ClassifiedAdChallenged is published when a moderator challenges a classified
// ad, asking the seller for corrections.
type ClassifiedAdChallenged struct {
	ClassifiedAdID string
	ModeratorID    string
	Reason         string
	OccurredAt     time.Time
}

// EventType returns the public event type name.
func (e *ClassifiedAdChallenged) EventType() string {
	return ClassifiedAdChallengedEventType
}

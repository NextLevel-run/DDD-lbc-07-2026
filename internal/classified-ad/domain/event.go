package domain

import (
	"time"

	"github.com/google/uuid"
)

// ClassifiedAdPostedEvent is emitted after a new ad has been posted and
// successfully persisted. The ad still awaits moderation (PendingReview) at this
// point. The aggregate is exposed so consumers can read whatever they need.
type ClassifiedAdPostedEvent struct {
	id           string    // Event ID (private)
	eventType    string    // Event type (private)
	emitedAt     time.Time // When the event was created (private)
	ClassifiedAd *ClassifiedAd
}

// NewClassifiedAdPostedEvent is the full constructor with all parameters.
func NewClassifiedAdPostedEvent(emitedAt time.Time, ad *ClassifiedAd) *ClassifiedAdPostedEvent {
	return &ClassifiedAdPostedEvent{
		id:           uuid.New().String(),
		eventType:    "ClassifiedAdPosted", // Past tense, no suffix
		emitedAt:     emitedAt,
		ClassifiedAd: ad,
	}
}

// NewClassifiedAdPostedEventFrom is a convenience constructor using the current time.
func NewClassifiedAdPostedEventFrom(ad *ClassifiedAd) *ClassifiedAdPostedEvent {
	return NewClassifiedAdPostedEvent(time.Now(), ad)
}

// EventType implements eventbus.DomainEvent.
func (e *ClassifiedAdPostedEvent) EventType() string {
	return e.eventType
}

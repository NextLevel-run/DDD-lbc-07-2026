package domain

import (
	"time"

	"github.com/google/uuid"
)

// ClassifiedAdPostedEvent is emitted when a seller posts a new classified ad
type ClassifiedAdPostedEvent struct {
	id           string
	eventType    string
	emitedAt     time.Time
	ClassifiedAd *ClassifiedAd
}

func NewClassifiedAdPostedEvent(emitedAt time.Time, classifiedAd *ClassifiedAd) *ClassifiedAdPostedEvent {
	return &ClassifiedAdPostedEvent{
		id:           uuid.New().String(),
		eventType:    "ClassifiedAdPosted",
		emitedAt:     emitedAt,
		ClassifiedAd: classifiedAd,
	}
}

func NewClassifiedAdPostedEventFrom(classifiedAd *ClassifiedAd) *ClassifiedAdPostedEvent {
	return NewClassifiedAdPostedEvent(time.Now(), classifiedAd)
}

func (e *ClassifiedAdPostedEvent) EventType() string {
	return e.eventType
}

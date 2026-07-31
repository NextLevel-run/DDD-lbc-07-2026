package domain

import "time"

// ModerationTaskClaimedEvent is emitted when a moderator claims a moderation task.
type ModerationTaskClaimedEvent struct {
	TaskID         string
	ClassifiedAdID string
	ModeratorID    string
	OccurredAt     time.Time
}

// EventType returns the event type name.
func (e *ModerationTaskClaimedEvent) EventType() string {
	return "ModerationTaskClaimed"
}

// NewModerationTaskClaimedEventFromTask builds a ModerationTaskClaimedEvent from
// a freshly claimed task (moderatorID and claimedAt are expected to be set).
func NewModerationTaskClaimedEventFromTask(task *ModerationTask) *ModerationTaskClaimedEvent {
	var moderatorID string
	if task.ModeratorID() != nil {
		moderatorID = task.ModeratorID().String()
	}
	var occurredAt time.Time
	if task.ClaimedAt() != nil {
		occurredAt = *task.ClaimedAt()
	}
	return &ModerationTaskClaimedEvent{
		TaskID:         task.ID().String(),
		ClassifiedAdID: task.ClassifiedAdID(),
		ModeratorID:    moderatorID,
		OccurredAt:     occurredAt,
	}
}

// ModerationTaskCompletedEvent is emitted when a claimed task is completed
// (accept, reject or challenge) and physically deleted.
type ModerationTaskCompletedEvent struct {
	TaskID         string
	ClassifiedAdID string
	ModeratorID    string
	OccurredAt     time.Time
}

// EventType returns the event type name.
func (e *ModerationTaskCompletedEvent) EventType() string {
	return "ModerationTaskCompleted"
}

// NewModerationTaskCompletedEventFromTask builds a ModerationTaskCompletedEvent
// from a claimed task about to be deleted.
func NewModerationTaskCompletedEventFromTask(task *ModerationTask, occurredAt time.Time) *ModerationTaskCompletedEvent {
	var moderatorID string
	if task.ModeratorID() != nil {
		moderatorID = task.ModeratorID().String()
	}
	return &ModerationTaskCompletedEvent{
		TaskID:         task.ID().String(),
		ClassifiedAdID: task.ClassifiedAdID(),
		ModeratorID:    moderatorID,
		OccurredAt:     occurredAt,
	}
}

// ClassifiedAdApprovedEvent is the internal event emitted when a moderator
// approves a classified ad. The publisher relays it as the public
// ClassifiedAdApproved integration event.
type ClassifiedAdApprovedEvent struct {
	ClassifiedAdID string
	ModeratorID    string
	OccurredAt     time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdApprovedEvent) EventType() string {
	return "ClassifiedAdApproved"
}

// NewClassifiedAdApprovedEventFromTask builds a ClassifiedAdApprovedEvent from
// the claimed task being accepted.
func NewClassifiedAdApprovedEventFromTask(task *ModerationTask, occurredAt time.Time) *ClassifiedAdApprovedEvent {
	var moderatorID string
	if task.ModeratorID() != nil {
		moderatorID = task.ModeratorID().String()
	}
	return &ClassifiedAdApprovedEvent{
		ClassifiedAdID: task.ClassifiedAdID(),
		ModeratorID:    moderatorID,
		OccurredAt:     occurredAt,
	}
}

// ClassifiedAdRejectedEvent is the internal event emitted when a moderator
// rejects a classified ad. The publisher relays it as the public
// ClassifiedAdRejected integration event.
type ClassifiedAdRejectedEvent struct {
	ClassifiedAdID string
	ModeratorID    string
	Reason         string
	OccurredAt     time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdRejectedEvent) EventType() string {
	return "ClassifiedAdRejected"
}

// NewClassifiedAdRejectedEventFromTask builds a ClassifiedAdRejectedEvent from
// the claimed task being rejected.
func NewClassifiedAdRejectedEventFromTask(task *ModerationTask, reason RejectReason, occurredAt time.Time) *ClassifiedAdRejectedEvent {
	var moderatorID string
	if task.ModeratorID() != nil {
		moderatorID = task.ModeratorID().String()
	}
	return &ClassifiedAdRejectedEvent{
		ClassifiedAdID: task.ClassifiedAdID(),
		ModeratorID:    moderatorID,
		Reason:         string(reason),
		OccurredAt:     occurredAt,
	}
}

// ClassifiedAdChallengedEvent is the internal event emitted when a moderator
// challenges a classified ad, asking the seller for corrections. The publisher
// relays it as the public ClassifiedAdChallenged integration event.
type ClassifiedAdChallengedEvent struct {
	ClassifiedAdID string
	ModeratorID    string
	Reason         string
	OccurredAt     time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdChallengedEvent) EventType() string {
	return "ClassifiedAdChallenged"
}

// NewClassifiedAdChallengedEventFromTask builds a ClassifiedAdChallengedEvent
// from the claimed task being challenged.
func NewClassifiedAdChallengedEventFromTask(task *ModerationTask, reason ChallengeReason, occurredAt time.Time) *ClassifiedAdChallengedEvent {
	var moderatorID string
	if task.ModeratorID() != nil {
		moderatorID = task.ModeratorID().String()
	}
	return &ClassifiedAdChallengedEvent{
		ClassifiedAdID: task.ClassifiedAdID(),
		ModeratorID:    moderatorID,
		Reason:         string(reason),
		OccurredAt:     occurredAt,
	}
}

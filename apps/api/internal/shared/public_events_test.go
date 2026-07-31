package shared

import (
	"testing"

	"ddd-second-hand-marketplace/pkg/eventbus"
)

// TestPublicEvents_EventTypes asserts that every public event implements
// eventbus.DomainEvent and returns its expected event type string, matching
// the exported constant.
func TestPublicEvents_EventTypes(t *testing.T) {
	cases := []struct {
		event    eventbus.DomainEvent
		expected string
	}{
		{&ClassifiedAdSubmitted{}, "ClassifiedAdSubmitted"},
		{&ClassifiedAdEdited{}, "ClassifiedAdEdited"},
		{&ClassifiedAdPublished{}, "ClassifiedAdPublished"},
		{&ClassifiedAdDeleted{}, "ClassifiedAdDeleted"},
		{&ClassifiedAdExpired{}, "ClassifiedAdExpired"},
		{&ClassifiedAdApproved{}, "ClassifiedAdApproved"},
		{&ClassifiedAdRejected{}, "ClassifiedAdRejected"},
		{&ClassifiedAdChallenged{}, "ClassifiedAdChallenged"},
	}

	for _, tc := range cases {
		if got := tc.event.EventType(); got != tc.expected {
			t.Errorf("EventType() = %q, want %q", got, tc.expected)
		}
	}
}

// TestPublicEventTypeConstants asserts the exported constants match the
// EventType() of the corresponding events, so publishers and consumers can
// rely on them for subscriptions.
func TestPublicEventTypeConstants(t *testing.T) {
	cases := []struct {
		constant string
		event    eventbus.DomainEvent
	}{
		{ClassifiedAdSubmittedEventType, &ClassifiedAdSubmitted{}},
		{ClassifiedAdEditedEventType, &ClassifiedAdEdited{}},
		{ClassifiedAdPublishedEventType, &ClassifiedAdPublished{}},
		{ClassifiedAdDeletedEventType, &ClassifiedAdDeleted{}},
		{ClassifiedAdExpiredEventType, &ClassifiedAdExpired{}},
		{ClassifiedAdApprovedEventType, &ClassifiedAdApproved{}},
		{ClassifiedAdRejectedEventType, &ClassifiedAdRejected{}},
		{ClassifiedAdChallengedEventType, &ClassifiedAdChallenged{}},
	}

	for _, tc := range cases {
		if tc.constant != tc.event.EventType() {
			t.Errorf("constant %q does not match EventType() %q", tc.constant, tc.event.EventType())
		}
	}
}

// TestPublicEventBus_IsUsableAsBus asserts a sync in-memory bus satisfies the
// PublicEventBus abstraction and routes public events by their type constant.
func TestPublicEventBus_IsUsableAsBus(t *testing.T) {
	var bus PublicEventBus = eventbus.NewSyncInMemoryEventBus()

	received := false
	err := bus.Subscribe(ClassifiedAdApprovedEventType, func(event eventbus.DomainEvent) error {
		if _, ok := event.(*ClassifiedAdApproved); !ok {
			t.Errorf("expected *ClassifiedAdApproved, got %T", event)
		}
		received = true
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}

	if err := bus.Publish(&ClassifiedAdApproved{ClassifiedAdID: "ad-1", ModeratorID: "mod-1"}); err != nil {
		t.Fatalf("Publish() error: %v", err)
	}
	if !received {
		t.Error("expected handler to receive the published public event")
	}
}

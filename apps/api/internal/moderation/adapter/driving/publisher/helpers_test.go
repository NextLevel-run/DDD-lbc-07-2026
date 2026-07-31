package publisher

import (
	"testing"

	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"
)

// publisherTestSetup wires a synchronous internal bus and a synchronous public
// bus, with a collector capturing everything published on the public bus for
// the given event type.
type publisherTestSetup struct {
	internalBus eventbus.Bus
	publicBus   eventbus.Bus
	collector   *eventbustesting.EventCollector
}

func newPublisherTestSetup(t *testing.T, publicEventType string) *publisherTestSetup {
	t.Helper()

	internalBus := eventbus.NewSyncInMemoryEventBus()
	publicBus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()
	require.NoError(t, publicBus.Subscribe(publicEventType, collector.EventHandler()))

	return &publisherTestSetup{
		internalBus: internalBus,
		publicBus:   publicBus,
		collector:   collector,
	}
}

// mockEvent is a DomainEvent with an arbitrary type string, used to verify that
// publishers ignore events of an unexpected concrete type.
type mockEvent struct {
	eventType string
}

func (e *mockEvent) EventType() string {
	return e.eventType
}

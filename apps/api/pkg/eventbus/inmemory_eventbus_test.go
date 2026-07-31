package eventbus

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEvent is a simple test implementation of DomainEvent
type testEvent struct {
	eventType string
	data      string
}

func (e *testEvent) EventType() string {
	return e.eventType
}

// testEventCollector collects events for verification in tests
type testEventCollector struct {
	events []DomainEvent
	mu     sync.Mutex
	wg     sync.WaitGroup
}

func newTestEventCollector() *testEventCollector {
	return &testEventCollector{
		events: make([]DomainEvent, 0),
	}
}

func (c *testEventCollector) handler() EventHandler {
	return func(evt DomainEvent) error {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.events = append(c.events, evt)
		c.wg.Done()
		return nil
	}
}

func (c *testEventCollector) expectEvents(n int) {
	c.wg.Add(n)
}

func (c *testEventCollector) waitForEvents(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (c *testEventCollector) getEvents() []DomainEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.events
}

func TestNewAsyncInMemoryEventBus(t *testing.T) {
	bus := NewAsyncInMemoryEventBus()
	require.NotNil(t, bus)
	assert.Equal(t, asyncMode, bus.mode)
	assert.NotNil(t, bus.handlers)
}

func TestNewSyncInMemoryEventBus(t *testing.T) {
	bus := NewSyncInMemoryEventBus()
	require.NotNil(t, bus)
	assert.Equal(t, syncMode, bus.mode)
	assert.NotNil(t, bus.handlers)
}

func TestAsyncEventBus_PublishAndSubscribe(t *testing.T) {
	// Given
	bus := NewAsyncInMemoryEventBus()
	collector := newTestEventCollector()
	collector.expectEvents(1)

	err := bus.Subscribe("TestEvent", collector.handler())
	require.NoError(t, err)

	event := &testEvent{
		eventType: "TestEvent",
		data:      "test data",
	}

	// When
	err = bus.Publish(event)

	// Then
	require.NoError(t, err)
	assert.True(t, collector.waitForEvents(200*time.Millisecond), "Expected event to be received")

	events := collector.getEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "TestEvent", events[0].EventType())
}

func TestSyncEventBus_PublishAndSubscribe(t *testing.T) {
	// Given
	bus := NewSyncInMemoryEventBus()
	collector := newTestEventCollector()
	collector.expectEvents(1) // Still need to set up WaitGroup for handler

	err := bus.Subscribe("TestEvent", collector.handler())
	require.NoError(t, err)

	event := &testEvent{
		eventType: "TestEvent",
		data:      "test data",
	}

	// When
	err = bus.Publish(event)

	// Then
	require.NoError(t, err)

	// With sync mode, events should be available immediately (no need to wait)
	events := collector.getEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "TestEvent", events[0].EventType())
}

func TestAsyncEventBus_MultipleHandlers(t *testing.T) {
	// Given
	bus := NewAsyncInMemoryEventBus()
	collector1 := newTestEventCollector()
	collector2 := newTestEventCollector()
	collector1.expectEvents(1)
	collector2.expectEvents(1)

	err := bus.Subscribe("TestEvent", collector1.handler())
	require.NoError(t, err)
	err = bus.Subscribe("TestEvent", collector2.handler())
	require.NoError(t, err)

	event := &testEvent{
		eventType: "TestEvent",
		data:      "test data",
	}

	// When
	err = bus.Publish(event)

	// Then
	require.NoError(t, err)
	assert.True(t, collector1.waitForEvents(200*time.Millisecond), "Expected event in collector1")
	assert.True(t, collector2.waitForEvents(200*time.Millisecond), "Expected event in collector2")

	events1 := collector1.getEvents()
	events2 := collector2.getEvents()
	require.Len(t, events1, 1)
	require.Len(t, events2, 1)
}

func TestSyncEventBus_MultipleHandlers(t *testing.T) {
	// Given
	bus := NewSyncInMemoryEventBus()
	collector1 := newTestEventCollector()
	collector2 := newTestEventCollector()
	collector1.expectEvents(1) // Still need to set up WaitGroup for handler
	collector2.expectEvents(1) // Still need to set up WaitGroup for handler

	err := bus.Subscribe("TestEvent", collector1.handler())
	require.NoError(t, err)
	err = bus.Subscribe("TestEvent", collector2.handler())
	require.NoError(t, err)

	event := &testEvent{
		eventType: "TestEvent",
		data:      "test data",
	}

	// When
	err = bus.Publish(event)

	// Then
	require.NoError(t, err)

	// With sync mode, both handlers should have been called immediately
	events1 := collector1.getEvents()
	events2 := collector2.getEvents()
	require.Len(t, events1, 1)
	require.Len(t, events2, 1)
}

func TestAsyncEventBus_HandlerError(t *testing.T) {
	// Given
	bus := NewAsyncInMemoryEventBus()
	handlerCalled := false
	var mu sync.Mutex

	errorHandler := func(evt DomainEvent) error {
		mu.Lock()
		handlerCalled = true
		mu.Unlock()
		return errors.New("handler error")
	}

	err := bus.Subscribe("TestEvent", errorHandler)
	require.NoError(t, err)

	event := &testEvent{
		eventType: "TestEvent",
		data:      "test data",
	}

	// When
	err = bus.Publish(event)

	// Then
	require.NoError(t, err, "Publish should not return error even if handler fails")

	// Wait for async handler to execute
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.True(t, handlerCalled, "Handler should have been called despite error")
	mu.Unlock()
}

func TestSyncEventBus_HandlerError(t *testing.T) {
	// Given
	bus := NewSyncInMemoryEventBus()
	handlerCalled := false

	errorHandler := func(evt DomainEvent) error {
		handlerCalled = true
		return errors.New("handler error")
	}

	err := bus.Subscribe("TestEvent", errorHandler)
	require.NoError(t, err)

	event := &testEvent{
		eventType: "TestEvent",
		data:      "test data",
	}

	// When
	err = bus.Publish(event)

	// Then
	require.NoError(t, err, "Publish should not return error even if handler fails")
	assert.True(t, handlerCalled, "Handler should have been called despite error")
}

func TestSyncEventBus_MultipleHandlersWithErrors(t *testing.T) {
	// Given
	bus := NewSyncInMemoryEventBus()
	handler1Called := false
	handler2Called := false
	handler3Called := false

	handler1 := func(evt DomainEvent) error {
		handler1Called = true
		return errors.New("handler 1 error")
	}

	handler2 := func(evt DomainEvent) error {
		handler2Called = true
		return nil
	}

	handler3 := func(evt DomainEvent) error {
		handler3Called = true
		return errors.New("handler 3 error")
	}

	err := bus.Subscribe("TestEvent", handler1)
	require.NoError(t, err)
	err = bus.Subscribe("TestEvent", handler2)
	require.NoError(t, err)
	err = bus.Subscribe("TestEvent", handler3)
	require.NoError(t, err)

	event := &testEvent{
		eventType: "TestEvent",
		data:      "test data",
	}

	// When
	err = bus.Publish(event)

	// Then
	require.NoError(t, err, "Publish should not return error")
	assert.True(t, handler1Called, "Handler 1 should have been called")
	assert.True(t, handler2Called, "Handler 2 should have been called")
	assert.True(t, handler3Called, "Handler 3 should have been called despite previous errors")
}

func TestEventBus_NoSubscribers(t *testing.T) {
	tests := []struct {
		name string
		bus  Bus
	}{
		{"Async", NewAsyncInMemoryEventBus()},
		{"Sync", NewSyncInMemoryEventBus()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &testEvent{
				eventType: "TestEvent",
				data:      "test data",
			}

			err := tt.bus.Publish(event)
			require.NoError(t, err, "Publish with no subscribers should not error")
		})
	}
}

func TestEventBus_DifferentEventTypes(t *testing.T) {
	tests := []struct {
		name string
		bus  Bus
	}{
		{"Async", NewAsyncInMemoryEventBus()},
		{"Sync", NewSyncInMemoryEventBus()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			collector1 := newTestEventCollector()
			collector2 := newTestEventCollector()

			// Set up WaitGroup for handlers
			collector1.expectEvents(1)
			collector2.expectEvents(1)

			err := tt.bus.Subscribe("Event1", collector1.handler())
			require.NoError(t, err)
			err = tt.bus.Subscribe("Event2", collector2.handler())
			require.NoError(t, err)

			event1 := &testEvent{eventType: "Event1", data: "data1"}
			event2 := &testEvent{eventType: "Event2", data: "data2"}

			// When
			err = tt.bus.Publish(event1)
			require.NoError(t, err)
			err = tt.bus.Publish(event2)
			require.NoError(t, err)

			// Then
			if tt.name == "Async" {
				assert.True(t, collector1.waitForEvents(200*time.Millisecond))
				assert.True(t, collector2.waitForEvents(200*time.Millisecond))
			}

			events1 := collector1.getEvents()
			events2 := collector2.getEvents()

			require.Len(t, events1, 1)
			require.Len(t, events2, 1)
			assert.Equal(t, "Event1", events1[0].EventType())
			assert.Equal(t, "Event2", events2[0].EventType())
		})
	}
}

func TestSyncEventBus_ExecutionOrder(t *testing.T) {
	// Given
	bus := NewSyncInMemoryEventBus()
	executionOrder := make([]int, 0)
	var mu sync.Mutex

	handler1 := func(evt DomainEvent) error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		return nil
	}

	handler2 := func(evt DomainEvent) error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		return nil
	}

	handler3 := func(evt DomainEvent) error {
		mu.Lock()
		executionOrder = append(executionOrder, 3)
		mu.Unlock()
		return nil
	}

	err := bus.Subscribe("TestEvent", handler1)
	require.NoError(t, err)
	err = bus.Subscribe("TestEvent", handler2)
	require.NoError(t, err)
	err = bus.Subscribe("TestEvent", handler3)
	require.NoError(t, err)

	event := &testEvent{
		eventType: "TestEvent",
		data:      "test data",
	}

	// When
	err = bus.Publish(event)

	// Then
	require.NoError(t, err)
	mu.Lock()
	assert.Equal(t, []int{1, 2, 3}, executionOrder, "Handlers should execute in subscription order")
	mu.Unlock()
}

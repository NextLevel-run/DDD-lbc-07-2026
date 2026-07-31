package consumer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// fakeClock is a settable implementation of domain.Clock for deterministic tests.
type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

// consumerTestSetup wires the real commands and in-memory repositories behind
// a synchronous public bus, so publishing an event runs the consumers inline.
type consumerTestSetup struct {
	publicBus     eventbus.Bus
	taskRepo      *inmemory.InMemoryModerationTaskRepository
	historyRepo   *inmemory.InMemoryClassifiedAdHistoryRepository
	createTask    command.CreateModerationTaskCommand
	appendHistory command.AppendHistoryEntryCommand
}

func newConsumerTestSetup(t *testing.T) *consumerTestSetup {
	t.Helper()

	taskRepo := inmemory.NewInMemoryModerationTaskRepository()
	historyRepo := inmemory.NewInMemoryClassifiedAdHistoryRepository()
	clock := &fakeClock{now: time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)}

	return &consumerTestSetup{
		publicBus:     eventbus.NewSyncInMemoryEventBus(),
		taskRepo:      taskRepo,
		historyRepo:   historyRepo,
		createTask:    command.BuildCreateModerationTaskCommand(taskRepo, clock),
		appendHistory: command.BuildAppendHistoryEntryCommand(historyRepo),
	}
}

// findHistory returns the stored history for the given ad, failing the test if absent.
func (s *consumerTestSetup) findHistory(t *testing.T, classifiedAdID string) *domain.ClassifiedAdHistory {
	t.Helper()

	history, err := s.historyRepo.FindByClassifiedAdID(classifiedAdID)
	require.NoError(t, err)
	return history
}

// mockEvent is a DomainEvent with an arbitrary type string, used to verify that
// consumers ignore events of an unexpected concrete type.
type mockEvent struct {
	eventType string
}

func (e *mockEvent) EventType() string {
	return e.eventType
}

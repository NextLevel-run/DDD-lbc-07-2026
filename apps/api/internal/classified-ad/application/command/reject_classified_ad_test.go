package command

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rejectTestSetup struct {
	repo           *fakeClassifiedAdRepository
	clock          *fakeClock
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        RejectClassifiedAdCommand
}

func setupRejectTest(t *testing.T) *rejectTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	// A single collector on both event types preserves the emission order.
	err := bus.Subscribe("ClassifiedAdRejected", collector.EventHandler())
	require.NoError(t, err)
	err = bus.Subscribe("ClassifiedAdDeleted", collector.EventHandler())
	require.NoError(t, err)

	return &rejectTestSetup{
		repo:           repo,
		clock:          clock,
		eventBus:       bus,
		eventCollector: collector,
		command:        BuildRejectClassifiedAdCommand(repo, clock, bus),
	}
}

func TestRejectClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupRejectTest(t)
	ad := seedSubmittedAd(t, setup.repo, setup.clock.Now())

	// When
	err := setup.command(RejectClassifiedAdCommandArgs{AdID: ad.ID().String()})

	// Then
	require.NoError(t, err, "Expected no error when rejecting a submitted ad")

	// Verify persistence: the rejected ad is automatically deleted
	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDeleted, stored.Status())
	require.NotNil(t, stored.DeletedAt())
	assert.Equal(t, setup.clock.Now(), *stored.DeletedAt())
	assert.Equal(t, domain.DeleteReasonRejected, stored.DeleteReason())
	assert.False(t, stored.IsOnline())

	// Verify event emission: rejected first, then deleted
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 2)

	rejectedEvent, ok := events[0].(*domain.ClassifiedAdRejectedEvent)
	require.True(t, ok, "Expected first event to be *ClassifiedAdRejectedEvent")
	assert.Equal(t, ad.ID().String(), rejectedEvent.AdID)
	assert.Equal(t, setup.clock.Now(), rejectedEvent.OccurredAt)

	deletedEvent, ok := events[1].(*domain.ClassifiedAdDeletedEvent)
	require.True(t, ok, "Expected second event to be *ClassifiedAdDeletedEvent")
	assert.Equal(t, ad.ID().String(), deletedEvent.AdID)
	assert.Equal(t, string(domain.DeleteReasonRejected), deletedEvent.Reason)
	assert.Equal(t, setup.clock.Now(), deletedEvent.DeletedAt)
}

func TestRejectClassifiedAdCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupAd       func(t *testing.T, setup *rejectTestSetup) string
		expectedError error
	}{
		{
			name: "InvalidAdIDFormat",
			setupAd: func(t *testing.T, setup *rejectTestSetup) string {
				return "not-a-uuid"
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AdNotFound",
			setupAd: func(t *testing.T, setup *rejectTestSetup) string {
				return uuid.New().String()
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AlreadyPublished",
			setupAd: func(t *testing.T, setup *rejectTestSetup) string {
				return seedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			expectedError: domain.ErrCannotReject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupRejectTest(t)
			adID := tt.setupAd(t, setup)

			// When
			err := setup.command(RejectClassifiedAdCommandArgs{AdID: adID})

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			assertNoEventsEmitted(t, setup.eventCollector)
		})
	}
}

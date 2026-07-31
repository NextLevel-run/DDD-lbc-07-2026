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

type approveTestSetup struct {
	repo           *fakeClassifiedAdRepository
	clock          *fakeClock
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        ApproveClassifiedAdCommand
}

func setupApproveTest(t *testing.T) *approveTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("ClassifiedAdApproved", collector.EventHandler())
	require.NoError(t, err)

	return &approveTestSetup{
		repo:           repo,
		clock:          clock,
		eventBus:       bus,
		eventCollector: collector,
		command:        BuildApproveClassifiedAdCommand(repo, clock, bus),
	}
}

func TestApproveClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupApproveTest(t)
	ad := seedSubmittedAd(t, setup.repo, setup.clock.Now())

	// When
	err := setup.command(ApproveClassifiedAdCommandArgs{AdID: ad.ID().String()})

	// Then
	require.NoError(t, err, "Expected no error when approving a submitted ad")

	// Verify persistence
	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusApproved, stored.Status())
	assert.True(t, stored.PublishedAt().IsZero(), "Expected publishedAt to remain unset until the ad is published")
	assert.False(t, stored.IsOnline(), "Expected an approved ad not to be online yet")

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "ClassifiedAdApproved", events[0].EventType())

	event, ok := events[0].(*domain.ClassifiedAdApprovedEvent)
	require.True(t, ok, "Expected event to be *ClassifiedAdApprovedEvent")
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, setup.clock.Now(), event.OccurredAt)
}

func TestApproveClassifiedAdCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupAd       func(t *testing.T, setup *approveTestSetup) string
		expectedError error
	}{
		{
			name: "InvalidAdIDFormat",
			setupAd: func(t *testing.T, setup *approveTestSetup) string {
				return "not-a-uuid"
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AdNotFound",
			setupAd: func(t *testing.T, setup *approveTestSetup) string {
				return uuid.New().String()
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AlreadyPublished",
			setupAd: func(t *testing.T, setup *approveTestSetup) string {
				return seedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			expectedError: domain.ErrCannotApprove,
		},
		{
			name: "Challenged",
			setupAd: func(t *testing.T, setup *approveTestSetup) string {
				ad := seedSubmittedAd(t, setup.repo, setup.clock.Now())
				require.NoError(t, ad.Challenge())
				require.NoError(t, setup.repo.Save(ad))
				return ad.ID().String()
			},
			expectedError: domain.ErrCannotApprove,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupApproveTest(t)
			adID := tt.setupAd(t, setup)

			// When
			err := setup.command(ApproveClassifiedAdCommandArgs{AdID: adID})

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			assertNoEventsEmitted(t, setup.eventCollector)
		})
	}
}

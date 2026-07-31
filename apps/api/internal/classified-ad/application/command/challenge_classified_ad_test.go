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

type challengeTestSetup struct {
	repo           *fakeClassifiedAdRepository
	clock          *fakeClock
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        ChallengeClassifiedAdCommand
}

func setupChallengeTest(t *testing.T) *challengeTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("ClassifiedAdChallenged", collector.EventHandler())
	require.NoError(t, err)

	return &challengeTestSetup{
		repo:           repo,
		clock:          clock,
		eventBus:       bus,
		eventCollector: collector,
		command:        BuildChallengeClassifiedAdCommand(repo, clock, bus),
	}
}

func TestChallengeClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupChallengeTest(t)
	ad := seedSubmittedAd(t, setup.repo, setup.clock.Now())

	// When
	err := setup.command(ChallengeClassifiedAdCommandArgs{AdID: ad.ID().String()})

	// Then
	require.NoError(t, err, "Expected no error when challenging a submitted ad")

	// Verify persistence
	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusChallenged, stored.Status())
	assert.False(t, stored.IsOnline(), "Expected a challenged ad not to be online")

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "ClassifiedAdChallenged", events[0].EventType())

	event, ok := events[0].(*domain.ClassifiedAdChallengedEvent)
	require.True(t, ok, "Expected event to be *ClassifiedAdChallengedEvent")
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, setup.clock.Now(), event.OccurredAt)
}

func TestChallengeClassifiedAdCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupAd       func(t *testing.T, setup *challengeTestSetup) string
		expectedError error
	}{
		{
			name: "InvalidAdIDFormat",
			setupAd: func(t *testing.T, setup *challengeTestSetup) string {
				return "not-a-uuid"
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AdNotFound",
			setupAd: func(t *testing.T, setup *challengeTestSetup) string {
				return uuid.New().String()
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AlreadyPublished",
			setupAd: func(t *testing.T, setup *challengeTestSetup) string {
				return seedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			expectedError: domain.ErrCannotChallenge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupChallengeTest(t)
			adID := tt.setupAd(t, setup)

			// When
			err := setup.command(ChallengeClassifiedAdCommandArgs{AdID: adID})

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			assertNoEventsEmitted(t, setup.eventCollector)
		})
	}
}

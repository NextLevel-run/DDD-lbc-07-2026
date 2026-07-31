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

type publishTestSetup struct {
	repo           *fakeClassifiedAdRepository
	clock          *fakeClock
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        PublishClassifiedAdCommand
}

func setupPublishTest(t *testing.T) *publishTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("ClassifiedAdPublished", collector.EventHandler())
	require.NoError(t, err)

	return &publishTestSetup{
		repo:           repo,
		clock:          clock,
		eventBus:       bus,
		eventCollector: collector,
		command:        BuildPublishClassifiedAdCommand(repo, clock, bus),
	}
}

// seedApprovedAd creates and persists a valid classified ad already approved
// by moderation (StatusApproved), returning it.
func seedApprovedAd(t *testing.T, repo *fakeClassifiedAdRepository, submittedAt time.Time) *domain.ClassifiedAd {
	t.Helper()

	ad := newSubmittedAd(t, submittedAt)
	require.NoError(t, ad.Approve())
	require.NoError(t, repo.Save(ad))
	return ad
}

func TestPublishClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupPublishTest(t)
	ad := seedApprovedAd(t, setup.repo, setup.clock.Now())

	// When
	err := setup.command(PublishClassifiedAdCommandArgs{AdID: ad.ID().String()})

	// Then
	require.NoError(t, err, "Expected no error when publishing an approved ad")

	// Verify persistence
	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPublished, stored.Status())
	assert.Equal(t, setup.clock.Now(), stored.PublishedAt())
	assert.True(t, stored.IsOnline(), "Expected a published ad to be online")

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "ClassifiedAdPublished", events[0].EventType())

	event, ok := events[0].(*domain.ClassifiedAdPublishedEvent)
	require.True(t, ok, "Expected event to be *ClassifiedAdPublishedEvent")
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, ad.Title(), event.Title)
	assert.Equal(t, string(ad.Category()), event.Category)
	assert.Equal(t, ad.Seller().Email().String(), event.SellerEmail)
	assert.Equal(t, ad.Seller().Pseudo(), event.SellerPseudo)
	assert.Equal(t, setup.clock.Now(), event.PublishedAt)
}

func TestPublishClassifiedAdCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupAd       func(t *testing.T, setup *publishTestSetup) string
		expectedError error
	}{
		{
			name: "InvalidAdIDFormat",
			setupAd: func(t *testing.T, setup *publishTestSetup) string {
				return "not-a-uuid"
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AdNotFound",
			setupAd: func(t *testing.T, setup *publishTestSetup) string {
				return uuid.New().String()
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "NotApproved_StillSubmitted",
			setupAd: func(t *testing.T, setup *publishTestSetup) string {
				return seedSubmittedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			expectedError: domain.ErrCannotPublish,
		},
		{
			name: "AlreadyPublished",
			setupAd: func(t *testing.T, setup *publishTestSetup) string {
				return seedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			expectedError: domain.ErrCannotPublish,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupPublishTest(t)
			adID := tt.setupAd(t, setup)

			// When
			err := setup.command(PublishClassifiedAdCommandArgs{AdID: adID})

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			assertNoEventsEmitted(t, setup.eventCollector)
		})
	}
}

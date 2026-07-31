package command

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expireTestSetup struct {
	repo           *fakeClassifiedAdRepository
	clock          *fakeClock
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        ExpireOutdatedAdsCommand
}

func setupExpireTest(t *testing.T) *expireTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("ClassifiedAdExpired", collector.EventHandler())
	require.NoError(t, err)

	return &expireTestSetup{
		repo:           repo,
		clock:          clock,
		eventBus:       bus,
		eventCollector: collector,
		command:        BuildExpireOutdatedAdsCommand(repo, clock, bus),
	}
}

func TestExpireOutdatedAdsCommand_ExpiresOnlyOutdatedAds(t *testing.T) {
	// Given
	setup := setupExpireTest(t)
	now := setup.clock.Now()

	outdatedAd := seedAd(t, setup.repo, now.Add(-2*domain.AdLifetime))
	recentAd := seedAd(t, setup.repo, now)

	// When
	count, err := setup.command()

	// Then
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Expected exactly one ad to be expired")

	// Verify persistence
	storedOutdated, err := setup.repo.FindByID(outdatedAd.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusExpired, storedOutdated.Status())
	require.NotNil(t, storedOutdated.ExpiredAt())
	assert.Equal(t, now, *storedOutdated.ExpiredAt())

	storedRecent, err := setup.repo.FindByID(recentAd.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPublished, storedRecent.Status(), "Expected the recent ad to remain published")

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "ClassifiedAdExpired", events[0].EventType())

	event, ok := events[0].(*domain.ClassifiedAdExpiredEvent)
	require.True(t, ok, "Expected event to be *ClassifiedAdExpiredEvent")
	assert.Equal(t, outdatedAd.ID().String(), event.AdID)
	assert.Equal(t, outdatedAd.Seller().Email().String(), event.SellerEmail)
	assert.Equal(t, now, event.ExpiredAt)
}

func TestExpireOutdatedAdsCommand_NoOutdatedAds(t *testing.T) {
	// Given
	setup := setupExpireTest(t)
	seedAd(t, setup.repo, setup.clock.Now())

	// When
	count, err := setup.command()

	// Then
	require.NoError(t, err)
	assert.Zero(t, count)
	assertNoEventsEmitted(t, setup.eventCollector)
}

package consumer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/clock"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/consumer"
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"
)

type approvedConsumerTestSetup struct {
	repo              *inmemory.InMemoryClassifiedAdRepository
	fixedClock        *clock.FixedClock
	internalBus       eventbus.Bus
	publicBus         eventbus.Bus
	internalCollector *eventbustesting.EventCollector
}

func setupApprovedConsumerTest(t *testing.T) *approvedConsumerTestSetup {
	t.Helper()

	repo := inmemory.NewInMemoryClassifiedAdRepository()
	fixedClock := clock.NewFixedClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	internalBus := eventbus.NewSyncInMemoryEventBus()
	publicBus := eventbus.NewSyncInMemoryEventBus()
	internalCollector := eventbustesting.NewEventCollector()

	require.NoError(t, internalBus.Subscribe("ClassifiedAdApproved", internalCollector.EventHandler()))

	approveClassifiedAd := command.BuildApproveClassifiedAdCommand(repo, fixedClock, internalBus)
	require.NoError(t, consumer.NewClassifiedAdApprovedConsumer(publicBus, approveClassifiedAd))

	return &approvedConsumerTestSetup{
		repo:              repo,
		fixedClock:        fixedClock,
		internalBus:       internalBus,
		publicBus:         publicBus,
		internalCollector: internalCollector,
	}
}

func TestClassifiedAdApprovedConsumer_ApprovesSubmittedAd(t *testing.T) {
	// Given
	setup := setupApprovedConsumerTest(t)
	ad := seedSubmittedAd(t, setup.repo, setup.fixedClock.Now())

	// When
	err := setup.publicBus.Publish(&shared.ClassifiedAdApproved{
		ClassifiedAdID: ad.ID().String(),
		ModeratorID:    "moderator-1",
		OccurredAt:     setup.fixedClock.Now(),
	})

	// Then
	require.NoError(t, err)

	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusApproved, stored.Status())

	// The internal ClassifiedAdApproved event was emitted (feeding the chained
	// internal consumer that publishes the ad).
	events := setup.internalCollector.GetEvents()
	require.Len(t, events, 1)
	internalEvent, ok := events[0].(*domain.ClassifiedAdApprovedEvent)
	require.True(t, ok, "Expected event to be *domain.ClassifiedAdApprovedEvent")
	assert.Equal(t, ad.ID().String(), internalEvent.AdID)
}

func TestClassifiedAdApprovedConsumer_IgnoresUnexpectedPayloadType(t *testing.T) {
	// Given
	setup := setupApprovedConsumerTest(t)
	ad := seedSubmittedAd(t, setup.repo, setup.fixedClock.Now())

	// When: an event with the right type name but the wrong payload type
	err := setup.publicBus.Publish(&mockEvent{eventType: shared.ClassifiedAdApprovedEventType})

	// Then: nothing happened
	require.NoError(t, err)

	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusSubmitted, stored.Status())
	assert.Empty(t, setup.internalCollector.GetEvents())
}

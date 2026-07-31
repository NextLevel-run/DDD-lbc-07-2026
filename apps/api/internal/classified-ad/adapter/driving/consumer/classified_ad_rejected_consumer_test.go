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

type rejectedConsumerTestSetup struct {
	repo              *inmemory.InMemoryClassifiedAdRepository
	fixedClock        *clock.FixedClock
	publicBus         eventbus.Bus
	internalCollector *eventbustesting.EventCollector
}

func setupRejectedConsumerTest(t *testing.T) *rejectedConsumerTestSetup {
	t.Helper()

	repo := inmemory.NewInMemoryClassifiedAdRepository()
	fixedClock := clock.NewFixedClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	internalBus := eventbus.NewSyncInMemoryEventBus()
	publicBus := eventbus.NewSyncInMemoryEventBus()
	internalCollector := eventbustesting.NewEventCollector()

	require.NoError(t, internalBus.Subscribe("ClassifiedAdRejected", internalCollector.EventHandler()))
	require.NoError(t, internalBus.Subscribe("ClassifiedAdDeleted", internalCollector.EventHandler()))

	rejectClassifiedAd := command.BuildRejectClassifiedAdCommand(repo, fixedClock, internalBus)
	require.NoError(t, consumer.NewClassifiedAdRejectedConsumer(publicBus, rejectClassifiedAd))

	return &rejectedConsumerTestSetup{
		repo:              repo,
		fixedClock:        fixedClock,
		publicBus:         publicBus,
		internalCollector: internalCollector,
	}
}

func TestClassifiedAdRejectedConsumer_RejectsAndDeletesSubmittedAd(t *testing.T) {
	// Given
	setup := setupRejectedConsumerTest(t)
	ad := seedSubmittedAd(t, setup.repo, setup.fixedClock.Now())

	// When
	err := setup.publicBus.Publish(&shared.ClassifiedAdRejected{
		ClassifiedAdID: ad.ID().String(),
		ModeratorID:    "moderator-1",
		Reason:         "suspect_price",
		OccurredAt:     setup.fixedClock.Now(),
	})

	// Then: the ad was rejected and automatically deleted
	require.NoError(t, err)

	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDeleted, stored.Status())
	assert.Equal(t, domain.DeleteReasonRejected, stored.DeleteReason())
	require.NotNil(t, stored.DeletedAt())
	assert.Equal(t, setup.fixedClock.Now(), *stored.DeletedAt())

	// Both internal events were emitted, rejected first
	events := setup.internalCollector.GetEvents()
	require.Len(t, events, 2)
	assert.IsType(t, &domain.ClassifiedAdRejectedEvent{}, events[0])
	assert.IsType(t, &domain.ClassifiedAdDeletedEvent{}, events[1])
}

func TestClassifiedAdRejectedConsumer_IgnoresUnexpectedPayloadType(t *testing.T) {
	// Given
	setup := setupRejectedConsumerTest(t)
	ad := seedSubmittedAd(t, setup.repo, setup.fixedClock.Now())

	// When: an event with the right type name but the wrong payload type
	err := setup.publicBus.Publish(&mockEvent{eventType: shared.ClassifiedAdRejectedEventType})

	// Then: nothing happened
	require.NoError(t, err)

	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusSubmitted, stored.Status())
	assert.Empty(t, setup.internalCollector.GetEvents())
}

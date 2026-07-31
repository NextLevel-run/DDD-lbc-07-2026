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
)

func TestClassifiedAdApprovedInternalConsumer_PublishesApprovedAd(t *testing.T) {
	// Given
	repo := inmemory.NewInMemoryClassifiedAdRepository()
	fixedClock := clock.NewFixedClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	internalBus := eventbus.NewSyncInMemoryEventBus()

	publishClassifiedAd := command.BuildPublishClassifiedAdCommand(repo, fixedClock, internalBus)
	require.NoError(t, consumer.NewClassifiedAdApprovedInternalConsumer(internalBus, publishClassifiedAd))

	ad := seedSubmittedAd(t, repo, fixedClock.Now())
	require.NoError(t, ad.Approve())
	require.NoError(t, repo.Save(ad))

	// When
	err := internalBus.Publish(&domain.ClassifiedAdApprovedEvent{
		AdID:       ad.ID().String(),
		OccurredAt: fixedClock.Now(),
	})

	// Then
	require.NoError(t, err)

	stored, err := repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPublished, stored.Status())
	assert.Equal(t, fixedClock.Now(), stored.PublishedAt())
	assert.True(t, stored.IsOnline())
}

func TestClassifiedAdApprovedInternalConsumer_IgnoresUnexpectedPayloadType(t *testing.T) {
	// Given
	repo := inmemory.NewInMemoryClassifiedAdRepository()
	fixedClock := clock.NewFixedClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	internalBus := eventbus.NewSyncInMemoryEventBus()

	publishClassifiedAd := command.BuildPublishClassifiedAdCommand(repo, fixedClock, internalBus)
	require.NoError(t, consumer.NewClassifiedAdApprovedInternalConsumer(internalBus, publishClassifiedAd))

	ad := seedSubmittedAd(t, repo, fixedClock.Now())
	require.NoError(t, ad.Approve())
	require.NoError(t, repo.Save(ad))

	// When: an event with the right type name but the wrong payload type
	err := internalBus.Publish(&mockEvent{eventType: "ClassifiedAdApproved"})

	// Then: nothing happened
	require.NoError(t, err)

	stored, err := repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusApproved, stored.Status())
}

func TestClassifiedAdApprovedConsumers_ChainFromPublicApprovalToPublication(t *testing.T) {
	// Full approval chain: public ClassifiedAdApproved → ApproveClassifiedAd
	// (internal event) → chained internal consumer → PublishClassifiedAd.

	// Given
	repo := inmemory.NewInMemoryClassifiedAdRepository()
	fixedClock := clock.NewFixedClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	internalBus := eventbus.NewSyncInMemoryEventBus()
	publicBus := eventbus.NewSyncInMemoryEventBus()

	approveClassifiedAd := command.BuildApproveClassifiedAdCommand(repo, fixedClock, internalBus)
	publishClassifiedAd := command.BuildPublishClassifiedAdCommand(repo, fixedClock, internalBus)
	require.NoError(t, consumer.NewClassifiedAdApprovedConsumer(publicBus, approveClassifiedAd))
	require.NoError(t, consumer.NewClassifiedAdApprovedInternalConsumer(internalBus, publishClassifiedAd))

	ad := seedSubmittedAd(t, repo, fixedClock.Now())

	// When
	err := publicBus.Publish(&shared.ClassifiedAdApproved{
		ClassifiedAdID: ad.ID().String(),
		ModeratorID:    "moderator-1",
		OccurredAt:     fixedClock.Now(),
	})

	// Then: the ad went submitted → approved → published in one chain
	require.NoError(t, err)

	stored, err := repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPublished, stored.Status())
	assert.Equal(t, fixedClock.Now(), stored.PublishedAt())
	assert.True(t, stored.IsOnline())
}

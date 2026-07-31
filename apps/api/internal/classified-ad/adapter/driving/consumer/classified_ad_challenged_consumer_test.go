package consumer_test

import (
	"strings"
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
	mailertesting "ddd-second-hand-marketplace/pkg/mailer/testing"

	"github.com/google/uuid"
)

type challengedConsumerTestSetup struct {
	repo       *inmemory.InMemoryClassifiedAdRepository
	fixedClock *clock.FixedClock
	publicBus  eventbus.Bus
	mailerSpy  *mailertesting.MailerSpy
}

func setupChallengedConsumerTest(t *testing.T) *challengedConsumerTestSetup {
	t.Helper()

	repo := inmemory.NewInMemoryClassifiedAdRepository()
	fixedClock := clock.NewFixedClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	internalBus := eventbus.NewSyncInMemoryEventBus()
	publicBus := eventbus.NewSyncInMemoryEventBus()
	mailerSpy := mailertesting.NewMailerSpy()

	challengeClassifiedAd := command.BuildChallengeClassifiedAdCommand(repo, fixedClock, internalBus)
	require.NoError(t, consumer.NewClassifiedAdChallengedConsumer(publicBus, challengeClassifiedAd, repo, mailerSpy))

	return &challengedConsumerTestSetup{
		repo:       repo,
		fixedClock: fixedClock,
		publicBus:  publicBus,
		mailerSpy:  mailerSpy,
	}
}

func TestClassifiedAdChallengedConsumer_ChallengesAdAndEmailsSeller(t *testing.T) {
	// Given
	setup := setupChallengedConsumerTest(t)
	ad := seedSubmittedAd(t, setup.repo, setup.fixedClock.Now())

	// When
	err := setup.publicBus.Publish(&shared.ClassifiedAdChallenged{
		ClassifiedAdID: ad.ID().String(),
		ModeratorID:    "moderator-1",
		Reason:         "price_to_verify",
		OccurredAt:     setup.fixedClock.Now(),
	})

	// Then: the ad is challenged
	require.NoError(t, err)

	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusChallenged, stored.Status())

	// And the seller received an email mentioning the ad and the reason
	sent := setup.mailerSpy.GetSentSimpleEmails()
	require.Len(t, sent, 1)
	email := sent[0]
	assert.Equal(t, "seller@example.com", email.To)
	assert.Equal(t, "no-reply@marketplace.local", email.From)
	assert.True(t, strings.Contains(email.Title, "Vélo hollandais"), "expected subject to mention the ad title, got %s", email.Title)
	assert.True(t, strings.Contains(email.Body, "seller-pseudo"), "expected body to mention the seller pseudo, got %s", email.Body)
	assert.True(t, strings.Contains(email.Body, "price_to_verify"), "expected body to mention the challenge reason, got %s", email.Body)
}

func TestClassifiedAdChallengedConsumer_UnknownAd_NoEmailSent(t *testing.T) {
	// Given
	setup := setupChallengedConsumerTest(t)

	// When: the challenged ad does not exist (the command fails, the sync bus
	// logs the handler error and Publish itself still returns nil)
	err := setup.publicBus.Publish(&shared.ClassifiedAdChallenged{
		ClassifiedAdID: uuid.New().String(),
		ModeratorID:    "moderator-1",
		Reason:         "price_to_verify",
		OccurredAt:     setup.fixedClock.Now(),
	})

	// Then: no email was sent
	require.NoError(t, err)
	assert.Empty(t, setup.mailerSpy.GetSentSimpleEmails())
}

func TestClassifiedAdChallengedConsumer_IgnoresUnexpectedPayloadType(t *testing.T) {
	// Given
	setup := setupChallengedConsumerTest(t)
	ad := seedSubmittedAd(t, setup.repo, setup.fixedClock.Now())

	// When: an event with the right type name but the wrong payload type
	err := setup.publicBus.Publish(&mockEvent{eventType: shared.ClassifiedAdChallengedEventType})

	// Then: nothing happened
	require.NoError(t, err)

	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusSubmitted, stored.Status())
	assert.Empty(t, setup.mailerSpy.GetSentSimpleEmails())
}

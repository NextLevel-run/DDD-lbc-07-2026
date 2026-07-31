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

type makeOfferTestSetup struct {
	repo           *fakeClassifiedAdRepository
	clock          *fakeClock
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        MakeOfferCommand
}

func setupMakeOfferTest(t *testing.T) *makeOfferTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("BuyerOfferMade", collector.EventHandler())
	require.NoError(t, err)

	return &makeOfferTestSetup{
		repo:           repo,
		clock:          clock,
		eventBus:       bus,
		eventCollector: collector,
		command:        BuildMakeOfferCommand(repo, clock, bus),
	}
}

// seedAd creates and persists a valid, published classified ad, returning it.
// The ad goes through the full moderation happy path (submitted → approved →
// published), with publishedAt set to submittedAt.
func seedAd(t *testing.T, repo *fakeClassifiedAdRepository, submittedAt time.Time) *domain.ClassifiedAd {
	t.Helper()

	ad := newSubmittedAd(t, submittedAt)
	require.NoError(t, ad.Approve())
	require.NoError(t, ad.Publish(submittedAt))

	require.NoError(t, repo.Save(ad))
	return ad
}

func validMakeOfferArgs(adID string) MakeOfferCommandArgs {
	return MakeOfferCommandArgs{
		AdID:          adID,
		BuyerEmail:    "Buyer@Example.com",
		BuyerPseudo:   "buyer-pseudo",
		AmountInCents: 12000,
		Message:       "Bonjour, ça vous intéresse ?",
	}
}

func TestMakeOfferCommand_Success(t *testing.T) {
	// Given
	setup := setupMakeOfferTest(t)
	ad := seedAd(t, setup.repo, setup.clock.Now())
	args := validMakeOfferArgs(ad.ID().String())

	// When
	err := setup.command(args)

	// Then
	require.NoError(t, err, "Expected no error when making a valid offer")

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "BuyerOfferMade", events[0].EventType())

	event, ok := events[0].(*domain.BuyerOfferMadeEvent)
	require.True(t, ok, "Expected event to be *BuyerOfferMadeEvent")
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, ad.Title(), event.AdTitle)
	assert.Equal(t, ad.Seller().Email().String(), event.SellerEmail)
	assert.Equal(t, "buyer@example.com", event.BuyerEmail)
	assert.Equal(t, args.BuyerPseudo, event.BuyerPseudo)
	assert.Equal(t, args.AmountInCents, event.Amount)
	assert.Equal(t, args.Message, event.Message)
	assert.Equal(t, setup.clock.Now(), event.OccurredAt)
}

func TestMakeOfferCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupAd       func(t *testing.T, setup *makeOfferTestSetup) string // returns AdID to use
		mutate        func(args *MakeOfferCommandArgs)
		expectedError error
	}{
		{
			name: "InvalidAdIDFormat",
			setupAd: func(t *testing.T, setup *makeOfferTestSetup) string {
				seedAd(t, setup.repo, setup.clock.Now())
				return "not-a-uuid"
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AdNotFound",
			setupAd: func(t *testing.T, setup *makeOfferTestSetup) string {
				return uuid.New().String()
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AdNotAvailable",
			setupAd: func(t *testing.T, setup *makeOfferTestSetup) string {
				ad := seedAd(t, setup.repo, setup.clock.Now().Add(-2*domain.AdLifetime))
				require.True(t, ad.Expire(setup.clock.Now()))
				return ad.ID().String()
			},
			expectedError: domain.ErrAdNotAvailable,
		},
		{
			name: "InvalidBuyerEmail",
			setupAd: func(t *testing.T, setup *makeOfferTestSetup) string {
				return seedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *MakeOfferCommandArgs) { a.BuyerEmail = "not-an-email" },
			expectedError: domain.ErrInvalidEmail,
		},
		{
			name: "EmptyMessage",
			setupAd: func(t *testing.T, setup *makeOfferTestSetup) string {
				return seedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *MakeOfferCommandArgs) { a.Message = "" },
			expectedError: domain.ErrEmptyOfferMessage,
		},
		{
			name: "MessageTooLong",
			setupAd: func(t *testing.T, setup *makeOfferTestSetup) string {
				return seedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *MakeOfferCommandArgs) { a.Message = repeatChar("a", 1001) },
			expectedError: domain.ErrOfferMessageTooLong,
		},
		{
			name: "NegativeAmount",
			setupAd: func(t *testing.T, setup *makeOfferTestSetup) string {
				return seedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *MakeOfferCommandArgs) { a.AmountInCents = -1 },
			expectedError: domain.ErrNegativeOfferAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupMakeOfferTest(t)
			adID := tt.setupAd(t, setup)
			args := validMakeOfferArgs(adID)
			if tt.mutate != nil {
				tt.mutate(&args)
			}

			// When
			err := setup.command(args)

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			assertNoEventsEmitted(t, setup.eventCollector)
		})
	}
}

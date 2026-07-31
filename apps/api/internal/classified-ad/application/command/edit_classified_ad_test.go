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

type editTestSetup struct {
	repo           *fakeClassifiedAdRepository
	hasher         fakePasswordHasher
	clock          *fakeClock
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        EditClassifiedAdCommand
}

func setupEditTest(t *testing.T) *editTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	hasher := fakePasswordHasher{}
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("ClassifiedAdEdited", collector.EventHandler())
	require.NoError(t, err)

	return &editTestSetup{
		repo:           repo,
		hasher:         hasher,
		clock:          clock,
		eventBus:       bus,
		eventCollector: collector,
		command:        BuildEditClassifiedAdCommand(repo, hasher, clock, bus),
	}
}

// seedChallengedAd creates and persists a valid classified ad challenged by
// moderation (StatusChallenged), returning it. The seller is
// seller@example.com with plaintext password "supersecret".
func seedChallengedAd(t *testing.T, repo *fakeClassifiedAdRepository, submittedAt time.Time) *domain.ClassifiedAd {
	t.Helper()

	ad := newSubmittedAd(t, submittedAt)
	require.NoError(t, ad.Challenge())
	require.NoError(t, repo.Save(ad))
	return ad
}

func validEditArgs(adID string) EditClassifiedAdCommandArgs {
	return EditClassifiedAdCommandArgs{
		AdID:         adID,
		Email:        "seller@example.com",
		Password:     "supersecret",
		Title:        "Vélo hollandais (prix corrigé)",
		Description:  "Très bon état, prix ajusté suite à la modération.",
		PriceInCents: 12000,
		ImageURLs:    []string{"http://img/1.jpg", "http://img/2.jpg"},
		Category:     string(domain.CategoryConsumerGoods),
		ZipCode:      "69001",
		CityName:     "Lyon",
	}
}

func TestEditClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupEditTest(t)
	ad := seedChallengedAd(t, setup.repo, setup.clock.Now())
	args := validEditArgs(ad.ID().String())

	// When
	err := setup.command(args)

	// Then
	require.NoError(t, err, "Expected no error when editing a challenged ad with valid credentials")

	// Verify persistence: content replaced, ad re-submitted for moderation
	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusSubmitted, stored.Status())
	assert.Equal(t, args.Title, stored.Title())
	assert.Equal(t, args.Description, stored.Description())
	assert.Equal(t, args.PriceInCents, stored.Price().AmountInCents())
	assert.Equal(t, args.ImageURLs, stored.ImageURLs())
	assert.Equal(t, domain.CategoryConsumerGoods, stored.Category())
	assert.Equal(t, args.ZipCode, stored.Location().ZipCode())
	assert.Equal(t, args.CityName, stored.Location().CityName())
	assert.False(t, stored.IsOnline(), "Expected a re-submitted ad not to be online")

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "ClassifiedAdEdited", events[0].EventType())

	event, ok := events[0].(*domain.ClassifiedAdEditedEvent)
	require.True(t, ok, "Expected event to be *ClassifiedAdEditedEvent")
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, args.Title, event.Title)
	assert.Equal(t, args.Description, event.Description)
	assert.Equal(t, args.PriceInCents, event.PriceInCents)
	assert.Equal(t, args.ImageURLs, event.ImageURLs)
	assert.Equal(t, args.Category, event.Category)
	assert.Equal(t, args.ZipCode, event.ZipCode)
	assert.Equal(t, args.CityName, event.CityName)
	assert.Equal(t, "seller@example.com", event.SellerEmail)
	assert.Equal(t, "seller-pseudo", event.SellerPseudo)
	assert.Equal(t, setup.clock.Now(), event.OccurredAt)
}

func TestEditClassifiedAdCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupAd       func(t *testing.T, setup *editTestSetup) string
		mutate        func(args *EditClassifiedAdCommandArgs)
		expectedError error
	}{
		{
			name: "InvalidAdIDFormat",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return "not-a-uuid"
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AdNotFound",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return uuid.New().String()
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "InvalidEmailFormat",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return seedChallengedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *EditClassifiedAdCommandArgs) { a.Email = "not-an-email" },
			expectedError: domain.ErrInvalidEmail,
		},
		{
			name: "WrongEmail",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return seedChallengedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *EditClassifiedAdCommandArgs) { a.Email = "someone-else@example.com" },
			expectedError: domain.ErrInvalidCredentials,
		},
		{
			name: "WrongPassword",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return seedChallengedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *EditClassifiedAdCommandArgs) { a.Password = "wrong-password" },
			expectedError: domain.ErrInvalidCredentials,
		},
		{
			name: "NotChallenged_StillSubmitted",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return seedSubmittedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			expectedError: domain.ErrCannotEdit,
		},
		{
			name: "NotChallenged_Published",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return seedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			expectedError: domain.ErrCannotEdit,
		},
		{
			name: "InvalidCategory",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return seedChallengedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *EditClassifiedAdCommandArgs) { a.Category = "not-a-category" },
			expectedError: domain.ErrInvalidCategory,
		},
		{
			name: "InvalidZipCode",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return seedChallengedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *EditClassifiedAdCommandArgs) { a.ZipCode = "1234" },
			expectedError: domain.ErrInvalidZipCode,
		},
		{
			name: "EmptyTitle",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return seedChallengedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *EditClassifiedAdCommandArgs) { a.Title = "" },
			expectedError: domain.ErrEmptyTitle,
		},
		{
			name: "NegativePrice",
			setupAd: func(t *testing.T, setup *editTestSetup) string {
				return seedChallengedAd(t, setup.repo, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *EditClassifiedAdCommandArgs) { a.PriceInCents = -1 },
			expectedError: domain.ErrNegativePrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupEditTest(t)
			adID := tt.setupAd(t, setup)
			args := validEditArgs(adID)
			if tt.mutate != nil {
				tt.mutate(&args)
			}

			// When
			err := setup.command(args)

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			assertNoEventsEmitted(t, setup.eventCollector)

			// Verify no mutation happened, when the ad exists and id was well-formed
			if adUUID, parseErr := uuid.Parse(adID); parseErr == nil {
				if stored, findErr := setup.repo.FindByID(adUUID); findErr == nil {
					assert.NotEqual(t, validEditArgs(adID).Title, stored.Title(), "Expected the ad content to remain untouched on error")
				}
			}
		})
	}
}

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

type submitTestSetup struct {
	repo           *fakeClassifiedAdRepository
	hasher         fakePasswordHasher
	clock          *fakeClock
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        SubmitClassifiedAdCommand
}

func setupSubmitTest(t *testing.T) *submitTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	hasher := fakePasswordHasher{}
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("ClassifiedAdPublished", collector.EventHandler())
	require.NoError(t, err)

	return &submitTestSetup{
		repo:           repo,
		hasher:         hasher,
		clock:          clock,
		eventBus:       bus,
		eventCollector: collector,
		command:        BuildSubmitClassifiedAdCommand(repo, hasher, clock, bus),
	}
}

func validSubmitArgs() SubmitClassifiedAdCommandArgs {
	return SubmitClassifiedAdCommandArgs{
		Title:          "Vélo hollandais",
		Description:    "Très bon état, peu servi.",
		PriceInCents:   15000,
		SellerEmail:    "Seller@Example.com",
		SellerPseudo:   "seller-pseudo",
		SellerPassword: "supersecret",
		ImageURLs:      []string{"http://img/1.jpg"},
		Category:       string(domain.CategoryConsumerGoods),
		ZipCode:        "75001",
		CityName:       "Paris",
	}
}

func assertNoAdsInRepository(t *testing.T, repo *fakeClassifiedAdRepository) {
	t.Helper()
	assert.Zero(t, repo.count(), "Expected no classified ads in repository")
}

func assertNoEventsEmitted(t *testing.T, collector *eventbustesting.EventCollector) {
	t.Helper()
	assert.Empty(t, collector.GetEvents(), "Expected no events to be emitted")
}

func TestSubmitClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupSubmitTest(t)
	args := validSubmitArgs()

	// When
	adID, err := setup.command(args)

	// Then
	require.NoError(t, err, "Expected no error when submitting a valid classified ad")
	assert.NotEmpty(t, adID, "Expected a classified ad id to be returned")

	// Verify persistence
	require.Equal(t, 1, setup.repo.count())
	adUUID, err := uuid.Parse(adID)
	require.NoError(t, err)
	stored, err := setup.repo.FindByID(adUUID)
	require.NoError(t, err)
	assert.Equal(t, args.Title, stored.Title())
	assert.Equal(t, args.Description, stored.Description())
	assert.Equal(t, args.PriceInCents, stored.Price().AmountInCents())
	assert.Equal(t, "seller@example.com", stored.Seller().Email().String())
	assert.Equal(t, domain.StatusPublished, stored.Status())
	assert.Equal(t, setup.clock.Now(), stored.PublishedAt())

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "ClassifiedAdPublished", events[0].EventType())

	event, ok := events[0].(*domain.ClassifiedAdPublishedEvent)
	require.True(t, ok, "Expected event to be *ClassifiedAdPublishedEvent")
	assert.Equal(t, adID, event.AdID)
	assert.Equal(t, args.Title, event.Title)
	assert.Equal(t, string(domain.CategoryConsumerGoods), event.Category)
	assert.Equal(t, "seller@example.com", event.SellerEmail)
	assert.Equal(t, args.SellerPseudo, event.SellerPseudo)
	assert.Equal(t, setup.clock.Now(), event.PublishedAt)
}

func TestSubmitClassifiedAdCommand_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		mutate        func(args *SubmitClassifiedAdCommandArgs)
		expectedError error
	}{
		{
			name:          "InvalidSellerEmail",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.SellerEmail = "not-an-email" },
			expectedError: domain.ErrInvalidEmail,
		},
		{
			name:          "PasswordTooShort",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.SellerPassword = "short" },
			expectedError: domain.ErrPasswordTooShort,
		},
		{
			name:          "EmptyPseudo",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.SellerPseudo = "" },
			expectedError: domain.ErrEmptyPseudo,
		},
		{
			name:          "PseudoTooLong",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.SellerPseudo = repeatChar("a", 31) },
			expectedError: domain.ErrPseudoTooLong,
		},
		{
			name:          "InvalidCategory",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.Category = "not-a-category" },
			expectedError: domain.ErrInvalidCategory,
		},
		{
			name:          "InvalidZipCode",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.ZipCode = "1234" },
			expectedError: domain.ErrInvalidZipCode,
		},
		{
			name:          "EmptyCityName",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.CityName = "" },
			expectedError: domain.ErrEmptyCityName,
		},
		{
			name:          "EmptyTitle",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.Title = "" },
			expectedError: domain.ErrEmptyTitle,
		},
		{
			name:          "TitleTooLong",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.Title = repeatChar("a", 101) },
			expectedError: domain.ErrTitleTooLong,
		},
		{
			name:          "EmptyDescription",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.Description = "" },
			expectedError: domain.ErrEmptyDescription,
		},
		{
			name:          "NegativePrice",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.PriceInCents = -1 },
			expectedError: domain.ErrNegativePrice,
		},
		{
			name: "TooManyImages",
			mutate: func(a *SubmitClassifiedAdCommandArgs) {
				a.ImageURLs = make([]string, 11)
				for i := range a.ImageURLs {
					a.ImageURLs[i] = "http://img/x.jpg"
				}
			},
			expectedError: domain.ErrTooManyImages,
		},
		{
			name:          "EmptyImageURL",
			mutate:        func(a *SubmitClassifiedAdCommandArgs) { a.ImageURLs = []string{""} },
			expectedError: domain.ErrEmptyImageURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupSubmitTest(t)
			args := validSubmitArgs()
			tt.mutate(&args)

			// When
			adID, err := setup.command(args)

			// Then
			require.ErrorIs(t, err, tt.expectedError)
			assert.Empty(t, adID)
			assertNoAdsInRepository(t, setup.repo)
			assertNoEventsEmitted(t, setup.eventCollector)
		})
	}
}

func repeatChar(s string, n int) string {
	out := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, s[0])
	}
	return string(out)
}

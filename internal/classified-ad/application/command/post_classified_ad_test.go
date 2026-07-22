package command

import (
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeHasher is a deterministic stub avoiding the cost of real bcrypt in tests.
type fakeHasher struct{}

func (fakeHasher) Hash(plainPassword string) (string, error) {
	return "hashed:" + plainPassword, nil
}

type testSetup struct {
	repo           domain.ClassifiedAdRepository
	eventCollector *eventbustesting.EventCollector
	command        PostClassifiedAdCommand
}

func setupTest(t *testing.T) *testSetup {
	t.Helper()

	repo := inmemory.NewInMemoryClassifiedAdRepository()
	eventBus := eventbus.NewSyncInMemoryEventBus() // sync => deterministic tests
	collector := eventbustesting.NewEventCollector()

	err := eventBus.Subscribe("ClassifiedAdPosted", collector.EventHandler())
	require.NoError(t, err, "failed to subscribe to events")

	command := BuildPostClassifiedAdCommand(repo, fakeHasher{}, eventBus)

	return &testSetup{
		repo:           repo,
		eventCollector: collector,
		command:        command,
	}
}

func validArgs() PostClassifiedAdCommandArgs {
	return PostClassifiedAdCommandArgs{
		Title:          "iPhone 12",
		Description:    "Barely used, 128GB",
		PriceInCents:   19900,
		Category:       string(domain.CategoryElectronics),
		PhotoURLs:      []string{"https://example.com/iphone.jpg"},
		Location:       "Paris",
		SellerEmail:    "seller@example.com",
		SellerNickname: "seller99",
		SellerPassword: "s3cr3t",
	}
}

func assertNoAdsInRepository(t *testing.T, repo domain.ClassifiedAdRepository) {
	t.Helper()
	ads, err := repo.FindAll(domain.FindAllFilters{})
	require.NoError(t, err)
	assert.Empty(t, ads, "expected no ad to be persisted")
}

func TestPostClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupTest(t)

	// When
	adID, err := setup.command(validArgs())

	// Then
	require.NoError(t, err)
	assert.NotEmpty(t, adID)

	// Verify persistence
	ad, err := setup.repo.GetById(adID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPendingReview, ad.Status())
	assert.Equal(t, "iPhone 12", ad.Title().String())

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "ClassifiedAdPosted", events[0].EventType())

	event, ok := events[0].(*domain.ClassifiedAdPostedEvent)
	require.True(t, ok)
	assert.Equal(t, adID, event.ClassifiedAd.ID())
	assert.Equal(t, domain.StatusPendingReview, event.ClassifiedAd.Status())
}

func TestPostClassifiedAdCommand_NeverStoresPasswordInClearText(t *testing.T) {
	// Given
	setup := setupTest(t)

	// When
	adID, err := setup.command(validArgs())

	// Then
	require.NoError(t, err)
	ad, err := setup.repo.GetById(adID)
	require.NoError(t, err)
	assert.NotEqual(t, "s3cr3t", ad.Seller().HashedPassword().String())
	assert.Equal(t, "hashed:s3cr3t", ad.Seller().HashedPassword().String())
}

func TestPostClassifiedAdCommand_ValidationErrors_NoSideEffects(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*PostClassifiedAdCommandArgs)
		expectedErr error
	}{
		{"EmptyTitle", func(a *PostClassifiedAdCommandArgs) { a.Title = "" }, domain.ErrEmptyTitle},
		{"EmptyDescription", func(a *PostClassifiedAdCommandArgs) { a.Description = "" }, domain.ErrEmptyDescription},
		{"NegativePrice", func(a *PostClassifiedAdCommandArgs) { a.PriceInCents = -1 }, domain.ErrNegativePrice},
		{"InvalidCategory", func(a *PostClassifiedAdCommandArgs) { a.Category = "Nope" }, domain.ErrInvalidCategory},
		{"InvalidPhoto", func(a *PostClassifiedAdCommandArgs) { a.PhotoURLs = []string{"nope"} }, domain.ErrInvalidPhotoURL},
		{"EmptyLocation", func(a *PostClassifiedAdCommandArgs) { a.Location = "" }, domain.ErrEmptyLocation},
		{"InvalidEmail", func(a *PostClassifiedAdCommandArgs) { a.SellerEmail = "nope" }, domain.ErrInvalidEmail},
		{"EmptyNickname", func(a *PostClassifiedAdCommandArgs) { a.SellerNickname = "" }, domain.ErrEmptyNickname},
		{"EmptyPassword", func(a *PostClassifiedAdCommandArgs) { a.SellerPassword = "" }, domain.ErrEmptyPassword},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupTest(t)
			args := validArgs()
			tt.mutate(&args)

			// When
			_, err := setup.command(args)

			// Then
			require.ErrorIs(t, err, tt.expectedErr)
			assertNoAdsInRepository(t, setup.repo)
			assert.Empty(t, setup.eventCollector.GetEvents())
		})
	}
}

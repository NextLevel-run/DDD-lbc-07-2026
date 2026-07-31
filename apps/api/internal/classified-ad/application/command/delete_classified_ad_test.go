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

const (
	deleteTestSellerEmail    = "seller@example.com"
	deleteTestSellerPassword = "supersecret"
)

type deleteTestSetup struct {
	repo           *fakeClassifiedAdRepository
	hasher         fakePasswordHasher
	clock          *fakeClock
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        DeleteClassifiedAdCommand
}

func setupDeleteTest(t *testing.T) *deleteTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	hasher := fakePasswordHasher{}
	clock := newFakeClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("ClassifiedAdDeleted", collector.EventHandler())
	require.NoError(t, err)

	return &deleteTestSetup{
		repo:           repo,
		hasher:         hasher,
		clock:          clock,
		eventBus:       bus,
		eventCollector: collector,
		command:        BuildDeleteClassifiedAdCommand(repo, hasher, clock, bus),
	}
}

// seedAdForDelete seeds an ad owned by a seller whose known plaintext password is
// deleteTestSellerPassword, using the given hasher so Delete's credential check works.
func seedAdForDelete(t *testing.T, repo *fakeClassifiedAdRepository, hasher domain.PasswordHasher, submittedAt time.Time) *domain.ClassifiedAd {
	t.Helper()

	email, err := domain.NewEmail(deleteTestSellerEmail)
	require.NoError(t, err)
	password, err := domain.NewPassword(deleteTestSellerPassword, hasher)
	require.NoError(t, err)
	seller, err := domain.NewSeller(email, "seller-pseudo", password)
	require.NoError(t, err)
	category, err := domain.NewCategory(string(domain.CategoryConsumerGoods))
	require.NoError(t, err)
	location, err := domain.NewLocation("75001", "Paris")
	require.NoError(t, err)

	ad, err := domain.NewClassifiedAd(
		"Vélo hollandais",
		"Très bon état, peu servi.",
		15000,
		seller,
		[]string{"http://img/1.jpg"},
		category,
		location,
		domain.NewSubmissionDate(submittedAt),
	)
	require.NoError(t, err)

	// Drive the ad through the moderation happy path so it is published,
	// matching the pre-moderation behavior these tests were written against.
	require.NoError(t, ad.Approve())
	require.NoError(t, ad.Publish(submittedAt))

	require.NoError(t, repo.Save(ad))
	return ad
}

func validDeleteArgs(adID string) DeleteClassifiedAdCommandArgs {
	return DeleteClassifiedAdCommandArgs{
		AdID:     adID,
		Email:    deleteTestSellerEmail,
		Password: deleteTestSellerPassword,
		Reason:   string(domain.DeleteReasonSold),
	}
}

func TestDeleteClassifiedAdCommand_Success(t *testing.T) {
	// Given
	setup := setupDeleteTest(t)
	ad := seedAdForDelete(t, setup.repo, setup.hasher, setup.clock.Now())
	args := validDeleteArgs(ad.ID().String())

	// When
	err := setup.command(args)

	// Then
	require.NoError(t, err, "Expected no error when deleting with valid credentials")

	// Verify persistence
	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDeleted, stored.Status())
	require.NotNil(t, stored.DeletedAt())
	assert.Equal(t, setup.clock.Now(), *stored.DeletedAt())
	assert.Equal(t, domain.DeleteReasonSold, stored.DeleteReason())

	// Verify event emission
	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "ClassifiedAdDeleted", events[0].EventType())

	event, ok := events[0].(*domain.ClassifiedAdDeletedEvent)
	require.True(t, ok, "Expected event to be *ClassifiedAdDeletedEvent")
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, string(domain.DeleteReasonSold), event.Reason)
	assert.Equal(t, setup.clock.Now(), event.DeletedAt)
}

func TestDeleteClassifiedAdCommand_Idempotent_SecondDeleteIsNoOp(t *testing.T) {
	// Given
	setup := setupDeleteTest(t)
	ad := seedAdForDelete(t, setup.repo, setup.hasher, setup.clock.Now())
	args := validDeleteArgs(ad.ID().String())

	require.NoError(t, setup.command(args))
	require.Len(t, setup.eventCollector.GetEvents(), 1, "Sanity check: first delete emitted one event")

	// When: delete again, even with wrong credentials - should be a no-op, no error
	secondArgs := args
	secondArgs.Email = "someone-else@example.com"
	secondArgs.Password = "wrong-password"

	// Then
	err := setup.command(secondArgs)
	require.NoError(t, err, "Expected idempotent no-op (nil error) on already-deleted ad")

	// No additional event should have been published
	assert.Len(t, setup.eventCollector.GetEvents(), 1, "Expected no additional event on idempotent delete")

	stored, err := setup.repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDeleted, stored.Status())
}

func TestDeleteClassifiedAdCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupAd       func(t *testing.T, setup *deleteTestSetup) string
		mutate        func(args *DeleteClassifiedAdCommandArgs)
		expectedError error
	}{
		{
			name: "InvalidAdIDFormat",
			setupAd: func(t *testing.T, setup *deleteTestSetup) string {
				seedAdForDelete(t, setup.repo, setup.hasher, setup.clock.Now())
				return "not-a-uuid"
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "AdNotFound",
			setupAd: func(t *testing.T, setup *deleteTestSetup) string {
				return uuid.New().String()
			},
			expectedError: domain.ErrClassifiedAdNotFound,
		},
		{
			name: "InvalidEmailFormat",
			setupAd: func(t *testing.T, setup *deleteTestSetup) string {
				return seedAdForDelete(t, setup.repo, setup.hasher, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *DeleteClassifiedAdCommandArgs) { a.Email = "not-an-email" },
			expectedError: domain.ErrInvalidEmail,
		},
		{
			name: "InvalidReason",
			setupAd: func(t *testing.T, setup *deleteTestSetup) string {
				return seedAdForDelete(t, setup.repo, setup.hasher, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *DeleteClassifiedAdCommandArgs) { a.Reason = "not-a-reason" },
			expectedError: domain.ErrInvalidDeleteReason,
		},
		{
			name: "WrongPassword",
			setupAd: func(t *testing.T, setup *deleteTestSetup) string {
				return seedAdForDelete(t, setup.repo, setup.hasher, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *DeleteClassifiedAdCommandArgs) { a.Password = "wrong-password" },
			expectedError: domain.ErrInvalidCredentials,
		},
		{
			name: "WrongEmail",
			setupAd: func(t *testing.T, setup *deleteTestSetup) string {
				return seedAdForDelete(t, setup.repo, setup.hasher, setup.clock.Now()).ID().String()
			},
			mutate:        func(a *DeleteClassifiedAdCommandArgs) { a.Email = "someone-else@example.com" },
			expectedError: domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			setup := setupDeleteTest(t)
			adID := tt.setupAd(t, setup)
			args := validDeleteArgs(adID)
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
					assert.Equal(t, domain.StatusPublished, stored.Status(), "Expected ad to remain untouched on error")
					assert.Nil(t, stored.DeletedAt())
				}
			}
		})
	}
}

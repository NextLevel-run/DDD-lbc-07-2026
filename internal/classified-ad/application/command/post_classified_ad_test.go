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

type testSetup struct {
	repo           domain.ClassifiedAdRepository
	eventBus       eventbus.Bus
	eventCollector *eventbustesting.EventCollector
	command        PostClassifiedAdCommand
}

func setupTest(t *testing.T) *testSetup {
	t.Helper()

	repo := inmemory.NewInMemoryClassifiedAdRepository()
	bus := eventbus.NewSyncInMemoryEventBus()
	collector := eventbustesting.NewEventCollector()

	err := bus.Subscribe("ClassifiedAdPosted", collector.EventHandler())
	require.NoError(t, err)

	cmd := BuildPostClassifiedAdCommand(repo, bus)

	return &testSetup{
		repo:           repo,
		eventBus:       bus,
		eventCollector: collector,
		command:        cmd,
	}
}

func validArgs() PostClassifiedAdCommandArgs {
	return PostClassifiedAdCommandArgs{
		SellerId:      "seller-123",
		Title:         "Vélo VTT",
		Description:   "Vélo en très bon état",
		PriceAmount:   15000,
		PriceCurrency: "EUR",
		Category:      "Vehicles",
		PhotoURLs:     []string{"https://example.com/photo1.jpg"},
	}
}

func TestPostClassifiedAdCommand_Success(t *testing.T) {
	setup := setupTest(t)

	id, err := setup.command(validArgs())

	require.NoError(t, err)
	assert.NotEmpty(t, id)

	classifiedAd, err := setup.repo.GetById(id)
	require.NoError(t, err)
	assert.Equal(t, domain.ClassifiedAdPublished, classifiedAd.Status())
	assert.Equal(t, "seller-123", classifiedAd.SellerId())

	events := setup.eventCollector.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "ClassifiedAdPosted", events[0].EventType())

	event, ok := events[0].(*domain.ClassifiedAdPostedEvent)
	require.True(t, ok)
	assert.Equal(t, id, event.ClassifiedAd.ID())
}

func TestPostClassifiedAdCommand_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		mutate        func(args *PostClassifiedAdCommandArgs)
		expectedError error
	}{
		{
			name:          "EmptySellerId",
			mutate:        func(a *PostClassifiedAdCommandArgs) { a.SellerId = "" },
			expectedError: domain.ErrEmptySellerId,
		},
		{
			name:          "EmptyTitle",
			mutate:        func(a *PostClassifiedAdCommandArgs) { a.Title = "" },
			expectedError: domain.ErrEmptyTitle,
		},
		{
			name:          "NegativePrice",
			mutate:        func(a *PostClassifiedAdCommandArgs) { a.PriceAmount = -100 },
			expectedError: domain.ErrNegativeAmount,
		},
		{
			name:          "InvalidCategory",
			mutate:        func(a *PostClassifiedAdCommandArgs) { a.Category = "Toys" },
			expectedError: domain.ErrInvalidCategory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := setupTest(t)
			args := validArgs()
			tt.mutate(&args)

			id, err := setup.command(args)

			require.ErrorIs(t, err, tt.expectedError)
			assert.Empty(t, id)
			assert.Empty(t, setup.eventCollector.GetEvents())
		})
	}
}

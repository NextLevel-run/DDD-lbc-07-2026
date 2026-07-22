package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMoney_RejectsNegativeAmount(t *testing.T) {
	_, err := NewMoney(-100, EUR)

	require.ErrorIs(t, err, ErrNegativeAmount)
}

func TestNewMoney_RejectsInvalidCurrency(t *testing.T) {
	_, err := NewMoney(100, Currency("USD"))

	require.ErrorIs(t, err, ErrInvalidCurrency)
}

func TestNewCategory_RejectsInvalidCategory(t *testing.T) {
	_, err := NewCategory("Toys")

	require.ErrorIs(t, err, ErrInvalidCategory)
}

func TestNewCategory_AcceptsValidCategory(t *testing.T) {
	category, err := NewCategory("Vehicles")

	require.NoError(t, err)
	assert.Equal(t, CategoryVehicles, category)
}

func TestNewPhoto_RejectsEmptyURL(t *testing.T) {
	_, err := NewPhoto("")

	require.ErrorIs(t, err, ErrEmptyPhotoURL)
}

func TestNewClassifiedAd_CreatesWithPublishedStatusAndVersion1(t *testing.T) {
	classifiedAd, err := NewClassifiedAd(
		"seller-123",
		"Vélo VTT",
		"Vélo en très bon état",
		15000,
		"EUR",
		"Vehicles",
		[]string{"https://example.com/photo1.jpg"},
	)

	require.NoError(t, err)
	assert.NotEmpty(t, classifiedAd.ID())
	assert.Equal(t, 1, classifiedAd.Version())
	assert.Equal(t, ClassifiedAdPublished, classifiedAd.Status())
	assert.Equal(t, "seller-123", classifiedAd.SellerId())
	assert.Equal(t, "Vélo VTT", classifiedAd.Title())
	assert.Equal(t, int64(15000), classifiedAd.Price().Amount())
	assert.Equal(t, EUR, classifiedAd.Price().Currency())
	assert.Equal(t, CategoryVehicles, classifiedAd.Category())
	assert.Len(t, classifiedAd.Photos(), 1)
}

func TestNewClassifiedAd_AcceptsNoPhotos(t *testing.T) {
	classifiedAd, err := NewClassifiedAd(
		"seller-123",
		"Vélo VTT",
		"Vélo en très bon état",
		15000,
		"EUR",
		"Vehicles",
		[]string{},
	)

	require.NoError(t, err)
	assert.Empty(t, classifiedAd.Photos())
}

func TestNewClassifiedAd_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		sellerId      string
		title         string
		description   string
		priceAmount   int64
		priceCurrency string
		category      string
		photoURLs     []string
		expectedError error
	}{
		{
			name:     "EmptySellerId",
			sellerId: "", title: "Vélo", description: "Bon vélo",
			priceAmount: 100, priceCurrency: "EUR", category: "Vehicles",
			expectedError: ErrEmptySellerId,
		},
		{
			name:     "EmptyTitle",
			sellerId: "seller-1", title: "", description: "Bon vélo",
			priceAmount: 100, priceCurrency: "EUR", category: "Vehicles",
			expectedError: ErrEmptyTitle,
		},
		{
			name:     "EmptyDescription",
			sellerId: "seller-1", title: "Vélo", description: "",
			priceAmount: 100, priceCurrency: "EUR", category: "Vehicles",
			expectedError: ErrEmptyDescription,
		},
		{
			name:     "NegativePrice",
			sellerId: "seller-1", title: "Vélo", description: "Bon vélo",
			priceAmount: -100, priceCurrency: "EUR", category: "Vehicles",
			expectedError: ErrNegativeAmount,
		},
		{
			name:     "InvalidCurrency",
			sellerId: "seller-1", title: "Vélo", description: "Bon vélo",
			priceAmount: 100, priceCurrency: "USD", category: "Vehicles",
			expectedError: ErrInvalidCurrency,
		},
		{
			name:     "InvalidCategory",
			sellerId: "seller-1", title: "Vélo", description: "Bon vélo",
			priceAmount: 100, priceCurrency: "EUR", category: "Toys",
			expectedError: ErrInvalidCategory,
		},
		{
			name:     "EmptyPhotoURL",
			sellerId: "seller-1", title: "Vélo", description: "Bon vélo",
			priceAmount: 100, priceCurrency: "EUR", category: "Vehicles",
			photoURLs:     []string{""},
			expectedError: ErrEmptyPhotoURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClassifiedAd(
				tt.sellerId, tt.title, tt.description,
				tt.priceAmount, tt.priceCurrency, tt.category, tt.photoURLs,
			)

			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}

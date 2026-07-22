package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// VALUE OBJECT TESTS
// ============================================

func TestNewTitle_RejectsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"OnlyWhitespace", "   "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTitle(tt.input)
			require.ErrorIs(t, err, ErrEmptyTitle)
		})
	}
}

func TestNewDescription_RejectsEmpty(t *testing.T) {
	_, err := NewDescription("   ")
	require.ErrorIs(t, err, ErrEmptyDescription)
}

func TestNewPrice_AllowsZero(t *testing.T) {
	price, err := NewPrice(0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), price.AmountInCents())
}

func TestNewPrice_RejectsNegative(t *testing.T) {
	_, err := NewPrice(-1)
	require.ErrorIs(t, err, ErrNegativePrice)
}

func TestNewCategory_AcceptsKnownValues(t *testing.T) {
	for _, c := range []Category{CategoryElectronics, CategoryFurniture, CategoryVehicles, CategoryClothing, CategoryOther} {
		category, err := NewCategory(string(c))
		require.NoError(t, err)
		assert.Equal(t, c, category)
	}
}

func TestNewCategory_RejectsUnknownValue(t *testing.T) {
	_, err := NewCategory("Weapons")
	require.ErrorIs(t, err, ErrInvalidCategory)
}

func TestNewPhoto_ValidatesURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"ValidHTTPS", "https://cdn.example.com/photo.jpg", false},
		{"ValidHTTP", "http://example.com/img.png", false},
		{"Empty", "", true},
		{"NoScheme", "example.com/photo.jpg", true},
		{"NotAURL", "not a url", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPhoto(tt.input)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidPhotoURL)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewLocation_RejectsEmpty(t *testing.T) {
	_, err := NewLocation("  ")
	require.ErrorIs(t, err, ErrEmptyLocation)
}

func TestNewEmail_ValidatesFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid", "seller@example.com", false},
		{"InvalidNoAt", "invalid-email", true},
		{"Empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEmail(tt.input)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidEmail)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewNickname_RejectsEmpty(t *testing.T) {
	_, err := NewNickname("")
	require.ErrorIs(t, err, ErrEmptyNickname)
}

func TestNewHashedPassword_RejectsEmpty(t *testing.T) {
	_, err := NewHashedPassword("")
	require.ErrorIs(t, err, ErrEmptyHashedPassword)
}

// ============================================
// AGGREGATE ROOT TESTS
// ============================================

func TestNewClassifiedAd_CreatesInPendingReviewWithVersion1(t *testing.T) {
	// When
	ad := createTestClassifiedAd(t)

	// Then
	assert.NotEmpty(t, ad.ID())
	assert.Equal(t, 1, ad.Version())
	assert.Equal(t, StatusPendingReview, ad.Status())
	assert.Equal(t, "iPhone 12", ad.Title().String())
	assert.Equal(t, int64(19900), ad.Price().AmountInCents())
	assert.Equal(t, CategoryElectronics, ad.Category())
	assert.Equal(t, "Paris", ad.Location().String())
	assert.Equal(t, "seller@example.com", ad.Seller().Email().String())
	assert.False(t, ad.PostedAt().IsZero())
}

func TestNewClassifiedAd_AllowsZeroPhotos(t *testing.T) {
	// When
	ad, err := NewClassifiedAd(
		"Free couch", "Come pick it up", 0, string(CategoryFurniture),
		nil, "Lyon", "seller@example.com", "seller99", "hashed-secret",
	)

	// Then
	require.NoError(t, err)
	assert.Empty(t, ad.Photos())
}

func TestNewClassifiedAd_AcceptsMultiplePhotos(t *testing.T) {
	// When
	ad, err := NewClassifiedAd(
		"Bike", "Good condition", 5000, string(CategoryOther),
		[]string{"https://example.com/1.jpg", "https://example.com/2.jpg"},
		"Lille", "seller@example.com", "seller99", "hashed-secret",
	)

	// Then
	require.NoError(t, err)
	assert.Len(t, ad.Photos(), 2)
}

func TestNewClassifiedAd_ValidatesInvariants(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*classifiedAdArgs)
		expectedErr error
	}{
		{"EmptyTitle", func(a *classifiedAdArgs) { a.title = "" }, ErrEmptyTitle},
		{"EmptyDescription", func(a *classifiedAdArgs) { a.description = "" }, ErrEmptyDescription},
		{"NegativePrice", func(a *classifiedAdArgs) { a.priceInCents = -50 }, ErrNegativePrice},
		{"InvalidCategory", func(a *classifiedAdArgs) { a.category = "Nope" }, ErrInvalidCategory},
		{"InvalidPhoto", func(a *classifiedAdArgs) { a.photoURLs = []string{"not-a-url"} }, ErrInvalidPhotoURL},
		{"EmptyLocation", func(a *classifiedAdArgs) { a.location = "" }, ErrEmptyLocation},
		{"InvalidEmail", func(a *classifiedAdArgs) { a.sellerEmail = "nope" }, ErrInvalidEmail},
		{"EmptyNickname", func(a *classifiedAdArgs) { a.sellerNickname = "" }, ErrEmptyNickname},
		{"EmptyHashedPassword", func(a *classifiedAdArgs) { a.sellerHashedPassword = "" }, ErrEmptyHashedPassword},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := validClassifiedAdArgs()
			tt.mutate(&args)

			// When
			_, err := NewClassifiedAd(
				args.title, args.description, args.priceInCents, args.category,
				args.photoURLs, args.location, args.sellerEmail, args.sellerNickname,
				args.sellerHashedPassword,
			)

			// Then
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

// ============================================
// TEST HELPERS
// ============================================

type classifiedAdArgs struct {
	title                string
	description          string
	priceInCents         int64
	category             string
	photoURLs            []string
	location             string
	sellerEmail          string
	sellerNickname       string
	sellerHashedPassword string
}

func validClassifiedAdArgs() classifiedAdArgs {
	return classifiedAdArgs{
		title:                "iPhone 12",
		description:          "Barely used, 128GB",
		priceInCents:         19900,
		category:             string(CategoryElectronics),
		photoURLs:            []string{"https://example.com/iphone.jpg"},
		location:             "Paris",
		sellerEmail:          "seller@example.com",
		sellerNickname:       "seller99",
		sellerHashedPassword: "hashed-secret",
	}
}

func createTestClassifiedAd(t *testing.T) *ClassifiedAd {
	t.Helper()
	args := validClassifiedAdArgs()
	ad, err := NewClassifiedAd(
		args.title, args.description, args.priceInCents, args.category,
		args.photoURLs, args.location, args.sellerEmail, args.sellerNickname,
		args.sellerHashedPassword,
	)
	require.NoError(t, err)
	return ad
}

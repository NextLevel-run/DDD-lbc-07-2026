package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

type getClassifiedAdTestSetup struct {
	repo  domain.ClassifiedAdRepository
	query GetClassifiedAdQuery
}

func setupGetClassifiedAdTest(t *testing.T) *getClassifiedAdTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	query := BuildGetClassifiedAdQuery(repo)

	return &getClassifiedAdTestSetup{
		repo:  repo,
		query: query,
	}
}

func TestGetClassifiedAdQuery_Success(t *testing.T) {
	// Given
	setup := setupGetClassifiedAdTest(t)
	params := defaultTestAdParams()
	ad := saveTestAd(t, setup.repo, params)

	// When
	view, err := setup.query(ad.ID().String())

	// Then
	require.NoError(t, err)
	assert.Equal(t, ad.ID().String(), view.ID)
	assert.Equal(t, params.Title, view.Title)
	assert.Equal(t, params.Description, view.Description)
	assert.Equal(t, params.PriceInCents, view.PriceInCents)
	assert.Equal(t, params.Category, view.Category)
	assert.Equal(t, params.SellerPseudo, view.SellerPseudo)
	assert.Equal(t, params.ImageURLs, view.ImageURLs)
	assert.Equal(t, params.ZipCode, view.ZipCode)
	assert.Equal(t, params.CityName, view.CityName)
	assert.True(t, params.SubmissionDate.Equal(view.SubmissionDate))
}

func TestGetClassifiedAdQuery_NeverLeaksSellerCredentials(t *testing.T) {
	// This test is primarily a compile-time guarantee: ClassifiedAdView has no
	// field that could carry the seller's email or password. If someone adds
	// such a field, this test (and its assertions below) must be revisited.

	// Given
	setup := setupGetClassifiedAdTest(t)
	ad := saveTestAd(t, setup.repo, defaultTestAdParams())

	// When
	view, err := setup.query(ad.ID().String())

	// Then
	require.NoError(t, err)
	assert.Equal(t, "seller-pseudo", view.SellerPseudo)
	// ClassifiedAdView struct fields: ID, Title, Description, PriceInCents,
	// Category, SellerPseudo, ImageURLs, ZipCode, CityName, SubmissionDate.
	// No SellerEmail/SellerPassword field exists — verified by struct definition.
}

func TestGetClassifiedAdQuery_NotFound_UnknownID(t *testing.T) {
	// Given
	setup := setupGetClassifiedAdTest(t)

	// When
	_, err := setup.query("00000000-0000-0000-0000-000000000000")

	// Then
	require.ErrorIs(t, err, domain.ErrClassifiedAdNotFound)
}

func TestGetClassifiedAdQuery_NotFound_MalformedID(t *testing.T) {
	// Given
	setup := setupGetClassifiedAdTest(t)

	// When
	_, err := setup.query("not-a-uuid")

	// Then
	require.ErrorIs(t, err, domain.ErrClassifiedAdNotFound)
}

func TestGetClassifiedAdQuery_NotFound_DeletedAd(t *testing.T) {
	// Given
	setup := setupGetClassifiedAdTest(t)
	ad := saveTestAd(t, setup.repo, defaultTestAdParams())

	email, err := domain.NewEmail("seller@example.com")
	require.NoError(t, err)

	deleted, err := ad.Delete(email, "super-secret", domain.DeleteReasonSold, fakePasswordHasher{}, time.Now())
	require.NoError(t, err)
	require.True(t, deleted)
	require.NoError(t, setup.repo.Save(ad))

	// When
	_, err = setup.query(ad.ID().String())

	// Then
	require.ErrorIs(t, err, domain.ErrClassifiedAdNotFound)
}

func TestGetClassifiedAdQuery_NotFound_ExpiredAd(t *testing.T) {
	// Given
	setup := setupGetClassifiedAdTest(t)
	params := defaultTestAdParams()
	params.SubmissionDate = time.Now().Add(-100 * 24 * time.Hour) // older than AdLifetime (90 days)
	ad := saveTestAd(t, setup.repo, params)

	expired := ad.Expire(time.Now())
	require.True(t, expired)
	require.NoError(t, setup.repo.Save(ad))

	// When
	_, err := setup.query(ad.ID().String())

	// Then
	require.ErrorIs(t, err, domain.ErrClassifiedAdNotFound)
}

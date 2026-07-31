package query

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

// fakePasswordHasher is a simple in-memory implementation of domain.PasswordHasher
// for tests: it "hashes" by prefixing the plaintext, and compares accordingly.
type fakePasswordHasher struct{}

var errFakeHashMismatch = errors.New("hash mismatch")

func (fakePasswordHasher) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

func (fakePasswordHasher) Compare(hash, plain string) error {
	if hash == "hashed:"+plain {
		return nil
	}
	return errFakeHashMismatch
}

// testAdParams holds overridable fields used to build a test ClassifiedAd.
type testAdParams struct {
	Title          string
	Description    string
	PriceInCents   int64
	SellerEmail    string
	SellerPseudo   string
	ImageURLs      []string
	Category       string
	ZipCode        string
	CityName       string
	SubmissionDate time.Time
}

func defaultTestAdParams() testAdParams {
	return testAdParams{
		Title:          "Vélo hollandais",
		Description:    "Vélo en excellent état, peu utilisé.",
		PriceInCents:   15000,
		SellerEmail:    "seller@example.com",
		SellerPseudo:   "seller-pseudo",
		ImageURLs:      []string{"https://example.com/image1.jpg"},
		Category:       string(domain.CategoryConsumerGoods),
		ZipCode:        "75001",
		CityName:       "Paris",
		SubmissionDate: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
	}
}

// buildTestAd builds a valid, published ClassifiedAd from the given params, without saving it.
func buildTestAd(t *testing.T, params testAdParams) *domain.ClassifiedAd {
	t.Helper()

	email, err := domain.NewEmail(params.SellerEmail)
	require.NoError(t, err)

	password, err := domain.NewPassword("super-secret", fakePasswordHasher{})
	require.NoError(t, err)

	seller, err := domain.NewSeller(email, params.SellerPseudo, password)
	require.NoError(t, err)

	category, err := domain.NewCategory(params.Category)
	require.NoError(t, err)

	location, err := domain.NewLocation(params.ZipCode, params.CityName)
	require.NoError(t, err)

	submissionDate := domain.NewSubmissionDate(params.SubmissionDate)

	ad, err := domain.NewClassifiedAd(
		params.Title,
		params.Description,
		params.PriceInCents,
		seller,
		params.ImageURLs,
		category,
		location,
		submissionDate,
	)
	require.NoError(t, err)

	return ad
}

// saveTestAd builds and saves a valid, published ClassifiedAd, returning it.
func saveTestAd(t *testing.T, repo domain.ClassifiedAdRepository, params testAdParams) *domain.ClassifiedAd {
	t.Helper()

	ad := buildTestAd(t, params)
	require.NoError(t, repo.Save(ad))
	return ad
}

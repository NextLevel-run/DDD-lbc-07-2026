package inmemory_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

// fakePasswordHasher is a simple in-memory implementation of domain.PasswordHasher for tests.
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

	// Drive the ad through the moderation happy path so it is published,
	// with publishedAt set to SubmissionDate (as before moderation existed).
	require.NoError(t, ad.Approve())
	require.NoError(t, ad.Publish(params.SubmissionDate))

	return ad
}

func saveTestAd(t *testing.T, repo domain.ClassifiedAdRepository, params testAdParams) *domain.ClassifiedAd {
	t.Helper()

	ad := buildTestAd(t, params)
	require.NoError(t, repo.Save(ad))
	return ad
}

func TestInMemoryClassifiedAdRepository_SaveAndFindByID(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	ad := saveTestAd(t, repo, defaultTestAdParams())

	found, err := repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, ad.ID(), found.ID())
	assert.Equal(t, ad.Title(), found.Title())
}

func TestInMemoryClassifiedAdRepository_Save_Update(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()
	ad := saveTestAd(t, repo, defaultTestAdParams())

	hasher := fakePasswordHasher{}
	email, err := domain.NewEmail(defaultTestAdParams().SellerEmail)
	require.NoError(t, err)
	reason, err := domain.NewDeleteReason(string(domain.DeleteReasonSold))
	require.NoError(t, err)

	deleted, err := ad.Delete(email, "super-secret", reason, hasher, time.Now())
	require.NoError(t, err)
	require.True(t, deleted)

	require.NoError(t, repo.Save(ad))

	found, err := repo.FindByID(ad.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDeleted, found.Status())
}

func TestInMemoryClassifiedAdRepository_FindByID_NotFound(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	_, err := repo.FindByID(uuid.New())
	assert.ErrorIs(t, err, domain.ErrClassifiedAdNotFound)
}

func TestInMemoryClassifiedAdRepository_FindExpirable(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	oldParams := defaultTestAdParams()
	oldParams.SubmissionDate = now.Add(-100 * 24 * time.Hour) // published well over 90 days ago
	expirableAd := saveTestAd(t, repo, oldParams)

	freshParams := defaultTestAdParams()
	freshParams.SubmissionDate = now.Add(-1 * 24 * time.Hour)
	saveTestAd(t, repo, freshParams)

	deletedParams := defaultTestAdParams()
	deletedParams.SubmissionDate = now.Add(-200 * 24 * time.Hour)
	deletedAd := saveTestAd(t, repo, deletedParams)
	hasher := fakePasswordHasher{}
	email, err := domain.NewEmail(deletedParams.SellerEmail)
	require.NoError(t, err)
	reason, err := domain.NewDeleteReason(string(domain.DeleteReasonSold))
	require.NoError(t, err)
	_, err = deletedAd.Delete(email, "super-secret", reason, hasher, now)
	require.NoError(t, err)
	require.NoError(t, repo.Save(deletedAd))

	expirable, err := repo.FindExpirable(now)
	require.NoError(t, err)
	require.Len(t, expirable, 1)
	assert.Equal(t, expirableAd.ID(), expirable[0].ID())
}

func TestInMemoryClassifiedAdRepository_Search_OnlineOnly(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	onlineAd := saveTestAd(t, repo, defaultTestAdParams())

	deletedParams := defaultTestAdParams()
	deletedAd := saveTestAd(t, repo, deletedParams)
	hasher := fakePasswordHasher{}
	email, err := domain.NewEmail(deletedParams.SellerEmail)
	require.NoError(t, err)
	reason, err := domain.NewDeleteReason(string(domain.DeleteReasonSold))
	require.NoError(t, err)
	_, err = deletedAd.Delete(email, "super-secret", reason, hasher, time.Now())
	require.NoError(t, err)
	require.NoError(t, repo.Save(deletedAd))

	t.Run("OnlineOnly true excludes non-online ads", func(t *testing.T) {
		results, err := repo.Search(domain.SearchCriteria{OnlineOnly: true})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, onlineAd.ID(), results[0].ID())
	})

	t.Run("OnlineOnly false includes non-online ads", func(t *testing.T) {
		results, err := repo.Search(domain.SearchCriteria{OnlineOnly: false})
		require.NoError(t, err)
		require.Len(t, results, 2)
	})
}

func TestInMemoryClassifiedAdRepository_Search_ByCategory(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	autoParams := defaultTestAdParams()
	autoParams.Category = string(domain.CategoryAuto)
	autoAd := saveTestAd(t, repo, autoParams)

	saveTestAd(t, repo, defaultTestAdParams()) // consumer_goods

	category := domain.CategoryAuto
	results, err := repo.Search(domain.SearchCriteria{Category: &category})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, autoAd.ID(), results[0].ID())
}

func TestInMemoryClassifiedAdRepository_Search_ByZipCode(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	lyonParams := defaultTestAdParams()
	lyonParams.ZipCode = "69001"
	lyonParams.CityName = "Lyon"
	lyonAd := saveTestAd(t, repo, lyonParams)

	saveTestAd(t, repo, defaultTestAdParams()) // Paris 75001

	zip := "69001"
	results, err := repo.Search(domain.SearchCriteria{ZipCode: &zip})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, lyonAd.ID(), results[0].ID())
}

func TestInMemoryClassifiedAdRepository_Search_ByCityName(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	lyonParams := defaultTestAdParams()
	lyonParams.ZipCode = "69001"
	lyonParams.CityName = "Lyon"
	lyonAd := saveTestAd(t, repo, lyonParams)

	saveTestAd(t, repo, defaultTestAdParams())

	city := "Lyon"
	results, err := repo.Search(domain.SearchCriteria{CityName: &city})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, lyonAd.ID(), results[0].ID())
}

func TestInMemoryClassifiedAdRepository_Search_ByPriceRange(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	cheapParams := defaultTestAdParams()
	cheapParams.PriceInCents = 1000
	saveTestAd(t, repo, cheapParams)

	midParams := defaultTestAdParams()
	midParams.PriceInCents = 5000
	midAd := saveTestAd(t, repo, midParams)

	expensiveParams := defaultTestAdParams()
	expensiveParams.PriceInCents = 100000
	saveTestAd(t, repo, expensiveParams)

	min := int64(2000)
	max := int64(10000)
	results, err := repo.Search(domain.SearchCriteria{MinPriceInCents: &min, MaxPriceInCents: &max})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, midAd.ID(), results[0].ID())
}

func TestInMemoryClassifiedAdRepository_Search_ByKeywords(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	matchParams := defaultTestAdParams()
	matchParams.Title = "Superbe Trottinette électrique"
	matchAd := saveTestAd(t, repo, matchParams)

	descMatchParams := defaultTestAdParams()
	descMatchParams.Title = "Objet divers"
	descMatchParams.Description = "Une trottinette pliable en bon état"
	descMatchAd := saveTestAd(t, repo, descMatchParams)

	saveTestAd(t, repo, defaultTestAdParams()) // no match

	keywords := "trottinette"
	results, err := repo.Search(domain.SearchCriteria{Keywords: &keywords})
	require.NoError(t, err)
	require.Len(t, results, 2)

	ids := []uuid.UUID{results[0].ID(), results[1].ID()}
	assert.Contains(t, ids, matchAd.ID())
	assert.Contains(t, ids, descMatchAd.ID())
}

func TestInMemoryClassifiedAdRepository_Search_SortAndPaginate(t *testing.T) {
	repo := inmemory.NewInMemoryClassifiedAdRepository()

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var ids []uuid.UUID
	for i, price := range []int64{300, 100, 200} {
		params := defaultTestAdParams()
		params.PriceInCents = price
		params.SubmissionDate = base.Add(time.Duration(i) * time.Hour)
		ad := saveTestAd(t, repo, params)
		ids = append(ids, ad.ID())
	}
	// ids[0] price=300 submitted first (oldest)
	// ids[1] price=100 submitted second
	// ids[2] price=200 submitted third (newest)

	t.Run("date_desc default", func(t *testing.T) {
		results, err := repo.Search(domain.SearchCriteria{})
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, ids[2], results[0].ID())
		assert.Equal(t, ids[1], results[1].ID())
		assert.Equal(t, ids[0], results[2].ID())
	})

	t.Run("date_asc", func(t *testing.T) {
		results, err := repo.Search(domain.SearchCriteria{SortBy: "date_asc"})
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, ids[0], results[0].ID())
		assert.Equal(t, ids[1], results[1].ID())
		assert.Equal(t, ids[2], results[2].ID())
	})

	t.Run("price_asc", func(t *testing.T) {
		results, err := repo.Search(domain.SearchCriteria{SortBy: "price_asc"})
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, ids[1], results[0].ID()) // price 100
		assert.Equal(t, ids[2], results[1].ID()) // price 200
		assert.Equal(t, ids[0], results[2].ID()) // price 300
	})

	t.Run("price_desc", func(t *testing.T) {
		results, err := repo.Search(domain.SearchCriteria{SortBy: "price_desc"})
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, ids[0], results[0].ID()) // price 300
		assert.Equal(t, ids[2], results[1].ID()) // price 200
		assert.Equal(t, ids[1], results[2].ID()) // price 100
	})

	t.Run("pagination", func(t *testing.T) {
		results, err := repo.Search(domain.SearchCriteria{SortBy: "date_asc", Limit: 2, Offset: 1})
		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.Equal(t, ids[1], results[0].ID())
		assert.Equal(t, ids[2], results[1].ID())
	})

	t.Run("offset beyond length", func(t *testing.T) {
		results, err := repo.Search(domain.SearchCriteria{Offset: 10})
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

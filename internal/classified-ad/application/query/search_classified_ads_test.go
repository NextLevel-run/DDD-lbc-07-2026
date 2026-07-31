package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

type searchClassifiedAdsTestSetup struct {
	repo  domain.ClassifiedAdRepository
	query SearchClassifiedAdsQuery
}

func setupSearchClassifiedAdsTest(t *testing.T) *searchClassifiedAdsTestSetup {
	t.Helper()

	repo := newFakeClassifiedAdRepository()
	query := BuildSearchClassifiedAdsQuery(repo)

	return &searchClassifiedAdsTestSetup{
		repo:  repo,
		query: query,
	}
}

func strPtr(s string) *string { return &s }
func i64Ptr(i int64) *int64   { return &i }

func assertIDsInResults(t *testing.T, results []ClassifiedAdListItemView, ids ...string) {
	t.Helper()
	for _, id := range ids {
		found := false
		for _, item := range results {
			if item.ID == id {
				found = true
				break
			}
		}
		assert.True(t, found, "expected ad %s to be in results", id)
	}
}

func assertIDsNotInResults(t *testing.T, results []ClassifiedAdListItemView, ids ...string) {
	t.Helper()
	for _, id := range ids {
		for _, item := range results {
			assert.NotEqual(t, id, item.ID, "did not expect ad %s to be in results", id)
		}
	}
}

func TestSearchClassifiedAdsQuery_FilterByCategory(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	autoParams := defaultTestAdParams()
	autoParams.Category = string(domain.CategoryAuto)
	autoAd := saveTestAd(t, setup.repo, autoParams)

	immoParams := defaultTestAdParams()
	immoParams.Category = string(domain.CategoryImmo)
	immoAd := saveTestAd(t, setup.repo, immoParams)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{Category: strPtr(string(domain.CategoryAuto))})

	// Then
	require.NoError(t, err)
	assertIDsInResults(t, results, autoAd.ID().String())
	assertIDsNotInResults(t, results, immoAd.ID().String())
}

func TestSearchClassifiedAdsQuery_InvalidCategory(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	// When
	_, err := setup.query(SearchClassifiedAdsQueryArgs{Category: strPtr("not-a-category")})

	// Then
	require.ErrorIs(t, err, domain.ErrInvalidCategory)
}

func TestSearchClassifiedAdsQuery_FilterByZipCode(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	parisParams := defaultTestAdParams()
	parisParams.ZipCode = "75001"
	parisParams.CityName = "Paris"
	parisAd := saveTestAd(t, setup.repo, parisParams)

	lyonParams := defaultTestAdParams()
	lyonParams.ZipCode = "69001"
	lyonParams.CityName = "Lyon"
	lyonAd := saveTestAd(t, setup.repo, lyonParams)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{ZipCode: strPtr("75001")})

	// Then
	require.NoError(t, err)
	assertIDsInResults(t, results, parisAd.ID().String())
	assertIDsNotInResults(t, results, lyonAd.ID().String())
}

func TestSearchClassifiedAdsQuery_FilterByCityName(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	parisParams := defaultTestAdParams()
	parisParams.CityName = "Paris"
	parisAd := saveTestAd(t, setup.repo, parisParams)

	lyonParams := defaultTestAdParams()
	lyonParams.CityName = "Lyon"
	lyonAd := saveTestAd(t, setup.repo, lyonParams)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{CityName: strPtr("Paris")})

	// Then
	require.NoError(t, err)
	assertIDsInResults(t, results, parisAd.ID().String())
	assertIDsNotInResults(t, results, lyonAd.ID().String())
}

func TestSearchClassifiedAdsQuery_FilterByPriceRange(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	cheapParams := defaultTestAdParams()
	cheapParams.PriceInCents = 1000
	cheapAd := saveTestAd(t, setup.repo, cheapParams)

	midParams := defaultTestAdParams()
	midParams.PriceInCents = 5000
	midAd := saveTestAd(t, setup.repo, midParams)

	expensiveParams := defaultTestAdParams()
	expensiveParams.PriceInCents = 10000
	expensiveAd := saveTestAd(t, setup.repo, expensiveParams)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{
		MinPriceInCents: i64Ptr(2000),
		MaxPriceInCents: i64Ptr(8000),
	})

	// Then
	require.NoError(t, err)
	assertIDsInResults(t, results, midAd.ID().String())
	assertIDsNotInResults(t, results, cheapAd.ID().String(), expensiveAd.ID().String())
}

func TestSearchClassifiedAdsQuery_FilterByKeywords_MatchesTitleOrDescription(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	titleMatchParams := defaultTestAdParams()
	titleMatchParams.Title = "Superbe VÉLO de course"
	titleMatchAd := saveTestAd(t, setup.repo, titleMatchParams)

	descriptionMatchParams := defaultTestAdParams()
	descriptionMatchParams.Title = "Objet divers"
	descriptionMatchParams.Description = "Ce vélo est en parfait état."
	descriptionMatchAd := saveTestAd(t, setup.repo, descriptionMatchParams)

	noMatchParams := defaultTestAdParams()
	noMatchParams.Title = "Canapé"
	noMatchParams.Description = "Canapé trois places."
	noMatchAd := saveTestAd(t, setup.repo, noMatchParams)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{Keywords: strPtr("vélo")})

	// Then
	require.NoError(t, err)
	assertIDsInResults(t, results, titleMatchAd.ID().String(), descriptionMatchAd.ID().String())
	assertIDsNotInResults(t, results, noMatchAd.ID().String())
}

func TestSearchClassifiedAdsQuery_SortByDateDesc_Default(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	oldParams := defaultTestAdParams()
	oldParams.SubmissionDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldAd := saveTestAd(t, setup.repo, oldParams)

	newParams := defaultTestAdParams()
	newParams.SubmissionDate = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	newAd := saveTestAd(t, setup.repo, newParams)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{})

	// Then
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, newAd.ID().String(), results[0].ID)
	assert.Equal(t, oldAd.ID().String(), results[1].ID)
}

func TestSearchClassifiedAdsQuery_SortByDateAsc(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	oldParams := defaultTestAdParams()
	oldParams.SubmissionDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldAd := saveTestAd(t, setup.repo, oldParams)

	newParams := defaultTestAdParams()
	newParams.SubmissionDate = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	newAd := saveTestAd(t, setup.repo, newParams)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{SortBy: "date_asc"})

	// Then
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, oldAd.ID().String(), results[0].ID)
	assert.Equal(t, newAd.ID().String(), results[1].ID)
}

func TestSearchClassifiedAdsQuery_SortByPriceAsc(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	cheapParams := defaultTestAdParams()
	cheapParams.PriceInCents = 1000
	cheapAd := saveTestAd(t, setup.repo, cheapParams)

	expensiveParams := defaultTestAdParams()
	expensiveParams.PriceInCents = 9000
	expensiveAd := saveTestAd(t, setup.repo, expensiveParams)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{SortBy: "price_asc"})

	// Then
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, cheapAd.ID().String(), results[0].ID)
	assert.Equal(t, expensiveAd.ID().String(), results[1].ID)
}

func TestSearchClassifiedAdsQuery_SortByPriceDesc(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	cheapParams := defaultTestAdParams()
	cheapParams.PriceInCents = 1000
	cheapAd := saveTestAd(t, setup.repo, cheapParams)

	expensiveParams := defaultTestAdParams()
	expensiveParams.PriceInCents = 9000
	expensiveAd := saveTestAd(t, setup.repo, expensiveParams)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{SortBy: "price_desc"})

	// Then
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, expensiveAd.ID().String(), results[0].ID)
	assert.Equal(t, cheapAd.ID().String(), results[1].ID)
}

func TestSearchClassifiedAdsQuery_PaginationDefaults(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	for i := 0; i < 25; i++ {
		params := defaultTestAdParams()
		params.SubmissionDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Hour)
		saveTestAd(t, setup.repo, params)
	}

	// When: no Limit/Offset/SortBy specified
	results, err := setup.query(SearchClassifiedAdsQueryArgs{})

	// Then: default limit of 20 applies
	require.NoError(t, err)
	assert.Len(t, results, 20)
}

func TestSearchClassifiedAdsQuery_Pagination_LimitAndOffset(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	var ads []*domain.ClassifiedAd
	for i := 0; i < 5; i++ {
		params := defaultTestAdParams()
		params.SubmissionDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Hour)
		ads = append(ads, saveTestAd(t, setup.repo, params))
	}
	// ads[4] is the most recent (date_desc default order puts it first)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{Limit: 2, Offset: 1})

	// Then
	require.NoError(t, err)
	require.Len(t, results, 2)
	// Expected order (date_desc): ads[4], ads[3], ads[2], ads[1], ads[0]
	// Offset 1, limit 2 -> ads[3], ads[2]
	assert.Equal(t, ads[3].ID().String(), results[0].ID)
	assert.Equal(t, ads[2].ID().String(), results[1].ID)
}

func TestSearchClassifiedAdsQuery_NeverReturnsNonOnlineAds(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)

	onlineAd := saveTestAd(t, setup.repo, defaultTestAdParams())

	deletedParams := defaultTestAdParams()
	deletedParams.SellerEmail = "deleted-seller@example.com"
	deletedAd := saveTestAd(t, setup.repo, deletedParams)
	email, err := domain.NewEmail("deleted-seller@example.com")
	require.NoError(t, err)
	deleted, err := deletedAd.Delete(email, "super-secret", domain.DeleteReasonSold, fakePasswordHasher{}, time.Now())
	require.NoError(t, err)
	require.True(t, deleted)
	require.NoError(t, setup.repo.Save(deletedAd))

	expiredParams := defaultTestAdParams()
	expiredParams.SubmissionDate = time.Now().Add(-100 * 24 * time.Hour)
	expiredAd := saveTestAd(t, setup.repo, expiredParams)
	expired := expiredAd.Expire(time.Now())
	require.True(t, expired)
	require.NoError(t, setup.repo.Save(expiredAd))

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{})

	// Then
	require.NoError(t, err)
	assertIDsInResults(t, results, onlineAd.ID().String())
	assertIDsNotInResults(t, results, deletedAd.ID().String(), expiredAd.ID().String())
}

func TestSearchClassifiedAdsQuery_MapsListItemViewFields(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)
	params := defaultTestAdParams()
	ad := saveTestAd(t, setup.repo, params)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{})

	// Then
	require.NoError(t, err)
	require.Len(t, results, 1)
	item := results[0]
	assert.Equal(t, ad.ID().String(), item.ID)
	assert.Equal(t, params.Title, item.Title)
	assert.Equal(t, params.PriceInCents, item.PriceInCents)
	assert.Equal(t, params.Category, item.Category)
	assert.Equal(t, params.CityName, item.CityName)
	assert.Equal(t, params.ZipCode, item.ZipCode)
	assert.Equal(t, params.ImageURLs[0], item.FirstImageURL)
	assert.True(t, params.SubmissionDate.Equal(item.SubmissionDate))
}

func TestSearchClassifiedAdsQuery_FirstImageURL_EmptyWhenNoImages(t *testing.T) {
	// Given
	setup := setupSearchClassifiedAdsTest(t)
	params := defaultTestAdParams()
	params.ImageURLs = []string{}
	ad := saveTestAd(t, setup.repo, params)

	// When
	results, err := setup.query(SearchClassifiedAdsQueryArgs{})

	// Then
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, ad.ID().String(), results[0].ID)
	assert.Equal(t, "", results[0].FirstImageURL)
}

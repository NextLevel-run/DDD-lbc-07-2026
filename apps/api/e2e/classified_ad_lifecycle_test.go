package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/bcrypt"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	httpadapter "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/http"
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/application/query"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// mutableClock is a domain.Clock whose current time can be advanced during a
// test, so expiration (a fixed 90-day lifetime) can be exercised without
// waiting for the real ticker.
type mutableClock struct {
	mu  sync.Mutex
	now time.Time
}

func newMutableClock(now time.Time) *mutableClock {
	return &mutableClock{now: now}
}

func (c *mutableClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *mutableClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// testServer bundles a running HTTP server with the pieces of the app needed
// to drive test-only actions that have no HTTP route, such as expiring ads.
type testServer struct {
	*httptest.Server
	clock             *mutableClock
	expireOutdatedAds command.ExpireOutdatedAdsCommand
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()

	repo := inmemory.NewInMemoryClassifiedAdRepository()
	hasher := bcrypt.NewBcryptPasswordHasher()
	testClock := newMutableClock(time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()

	submit := command.BuildSubmitClassifiedAdCommand(repo, hasher, testClock, bus)
	makeOffer := command.BuildMakeOfferCommand(repo, testClock, bus)
	deleteAd := command.BuildDeleteClassifiedAdCommand(repo, hasher, testClock, bus)
	expireOutdatedAds := command.BuildExpireOutdatedAdsCommand(repo, testClock, bus)
	search := query.BuildSearchClassifiedAdsQuery(repo)
	get := query.BuildGetClassifiedAdQuery(repo)

	handler := httpadapter.NewHandler(submit, makeOffer, deleteAd, search, get)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return &testServer{Server: srv, clock: testClock, expireOutdatedAds: expireOutdatedAds}
}

func (s *testServer) submitAd(t *testing.T, email, password string) string {
	t.Helper()

	body := httpadapter.SubmitClassifiedAdRequest{
		Title:          "Vélo hollandais",
		Description:    "Très bon état",
		PriceInCents:   15000,
		SellerEmail:    email,
		SellerPseudo:   "seller1",
		SellerPassword: password,
		ImageURLs:      []string{"http://example.com/img1.jpg"},
		Category:       "consumer_goods",
		ZipCode:        "75001",
		CityName:       "Paris",
	}
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := http.Post(s.URL+"/classified-ads", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody httpadapter.SubmitClassifiedAdResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody))
	require.NotEmpty(t, respBody.ID)

	return respBody.ID
}

func (s *testServer) get(t *testing.T, path string) *http.Response {
	t.Helper()

	resp, err := http.Get(s.URL + path)
	require.NoError(t, err)
	return resp
}

func (s *testServer) doJSON(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(method, s.URL+path, bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// TestClassifiedAdLifecycle drives the ClassifiedAd bounded context end to
// end through its real HTTP API: submit, search, get, make an offer, then
// delete.
func TestClassifiedAdLifecycle(t *testing.T) {
	srv := newTestServer(t)

	sellerEmail := "seller@example.com"
	sellerPassword := "supersecret"

	// 1. Submit the ad.
	id := srv.submitAd(t, sellerEmail, sellerPassword)

	// 2. It shows up in search.
	searchResp := srv.get(t, "/classified-ads?category=consumer_goods")
	defer searchResp.Body.Close()
	require.Equal(t, http.StatusOK, searchResp.StatusCode)

	var searchBody httpadapter.SearchClassifiedAdsResponse
	require.NoError(t, json.NewDecoder(searchResp.Body).Decode(&searchBody))
	require.Len(t, searchBody.Items, 1)
	assert.Equal(t, id, searchBody.Items[0].ID)

	// 3. It can be fetched by id.
	getResp := srv.get(t, "/classified-ads/"+id)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var getBody httpadapter.ClassifiedAdViewResponse
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&getBody))
	assert.Equal(t, id, getBody.ID)
	assert.Equal(t, "Vélo hollandais", getBody.Title)

	// 4. A buyer can make an offer.
	offerResp := srv.doJSON(t, http.MethodPost, "/classified-ads/"+id+"/offers", httpadapter.MakeOfferRequest{
		BuyerEmail:    "buyer@example.com",
		BuyerPseudo:   "buyer1",
		AmountInCents: 12000,
		Message:       "Intéressé !",
	})
	defer offerResp.Body.Close()
	assert.Equal(t, http.StatusCreated, offerResp.StatusCode)

	// 5. The seller deletes the ad.
	deleteResp := srv.doJSON(t, http.MethodDelete, "/classified-ads/"+id, httpadapter.DeleteClassifiedAdRequest{
		Email:    sellerEmail,
		Password: sellerPassword,
		Reason:   "sold",
	})
	defer deleteResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// 6. It is no longer reachable through the API.
	getAfterDeleteResp := srv.get(t, "/classified-ads/"+id)
	defer getAfterDeleteResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getAfterDeleteResp.StatusCode)

	searchAfterDeleteResp := srv.get(t, "/classified-ads?category=consumer_goods")
	defer searchAfterDeleteResp.Body.Close()
	require.Equal(t, http.StatusOK, searchAfterDeleteResp.StatusCode)

	var searchAfterDeleteBody httpadapter.SearchClassifiedAdsResponse
	require.NoError(t, json.NewDecoder(searchAfterDeleteResp.Body).Decode(&searchAfterDeleteBody))
	assert.Empty(t, searchAfterDeleteBody.Items)
}

// TestClassifiedAdLifecycle_Expiration exercises the other terminal
// transition: an ad that is never deleted becomes unreachable once its
// AdLifetime has elapsed and the expiration sweep runs. Expiration has no
// HTTP route (it is triggered by an internal ticker), so the test invokes
// the command directly to simulate the sweep after advancing the clock.
func TestClassifiedAdLifecycle_Expiration(t *testing.T) {
	srv := newTestServer(t)

	id := srv.submitAd(t, "seller@example.com", "supersecret")

	// Still reachable before the ad's lifetime has elapsed.
	beforeExpireCount, err := srv.expireOutdatedAds()
	require.NoError(t, err)
	assert.Equal(t, 0, beforeExpireCount)

	getBeforeResp := srv.get(t, "/classified-ads/"+id)
	defer getBeforeResp.Body.Close()
	assert.Equal(t, http.StatusOK, getBeforeResp.StatusCode)

	// Advance past the ad's lifetime and run the expiration sweep.
	srv.clock.Advance(domain.AdLifetime + time.Hour)

	expiredCount, err := srv.expireOutdatedAds()
	require.NoError(t, err)
	assert.Equal(t, 1, expiredCount)

	getAfterResp := srv.get(t, "/classified-ads/"+id)
	defer getAfterResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getAfterResp.StatusCode)

	searchResp := srv.get(t, "/classified-ads?category=consumer_goods")
	defer searchResp.Body.Close()
	require.Equal(t, http.StatusOK, searchResp.StatusCode)

	var searchBody httpadapter.SearchClassifiedAdsResponse
	require.NoError(t, json.NewDecoder(searchResp.Body).Decode(&searchBody))
	assert.Empty(t, searchBody.Items)
}

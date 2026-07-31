package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adhttp "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/http"
	addomain "ddd-second-hand-marketplace/internal/classified-ad/domain"
	moddomain "ddd-second-hand-marketplace/internal/moderation/domain"
)

// TestClassifiedAdLifecycle drives the ClassifiedAd bounded context end to
// end through its real HTTP API: submit, moderation approval (ads are born
// submitted and must be approved before becoming publicly visible), search,
// get, make an offer, then delete.
func TestClassifiedAdLifecycle(t *testing.T) {
	srv := newTestServer(t)

	sellerEmail := "seller@example.com"
	sellerPassword := "supersecret"

	// 1. Submit the ad: it awaits moderation and is not publicly visible yet.
	id := srv.submitAd(t, sellerEmail, sellerPassword)
	srv.assertPubliclyInvisible(t, id)

	// 2. A moderator approves it; the approved → published transition is
	// chained synchronously, so the ad comes online immediately.
	srv.approveAd(t, id)

	// 3. It shows up in search.
	searchBody := srv.searchAds(t)
	require.Len(t, searchBody.Items, 1)
	assert.Equal(t, id, searchBody.Items[0].ID)

	// 4. It can be fetched by id.
	getResp := srv.get(t, "/classified-ads/"+id)
	require.Equal(t, http.StatusOK, getResp.StatusCode)
	getBody := decodeJSON[adhttp.ClassifiedAdViewResponse](t, getResp)
	assert.Equal(t, id, getBody.ID)
	assert.Equal(t, "Vélo hollandais", getBody.Title)

	// 5. A buyer can make an offer.
	offerResp := srv.doJSON(t, http.MethodPost, "/classified-ads/"+id+"/offers", adhttp.MakeOfferRequest{
		BuyerEmail:    "buyer@example.com",
		BuyerPseudo:   "buyer1",
		AmountInCents: 12000,
		Message:       "Intéressé !",
	})
	defer offerResp.Body.Close()
	assert.Equal(t, http.StatusCreated, offerResp.StatusCode)

	// 6. The seller deletes the ad.
	deleteResp := srv.doJSON(t, http.MethodDelete, "/classified-ads/"+id, adhttp.DeleteClassifiedAdRequest{
		Email:    sellerEmail,
		Password: sellerPassword,
		Reason:   "sold",
	})
	defer deleteResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// 7. It is no longer reachable through the API.
	srv.assertPubliclyInvisible(t, id)

	// 8. The moderation history recorded the whole life of the ad.
	assert.Equal(t, []string{"submitted", "approved", "published", "deleted"}, srv.historyActions(t, id))
	entries := srv.historyEntries(t, id)
	require.NotNil(t, entries[3].Reason())
	assert.Equal(t, "sold", *entries[3].Reason())
}

// TestClassifiedAdLifecycle_Expiration exercises the last e2e journey of the
// spec: an approved and published ad that is never deleted becomes
// unreachable once its AdLifetime has elapsed and the expiration sweep runs,
// and the moderation history gains an "expired" entry. Expiration has no HTTP
// route (it is triggered by an internal ticker), so the test invokes the
// command directly to simulate the sweep after advancing the clock.
func TestClassifiedAdLifecycle_Expiration(t *testing.T) {
	srv := newTestServer(t)

	id := srv.submitAd(t, "seller@example.com", "supersecret")

	// The 90-day lifetime only starts at publication: get the ad approved and
	// published first.
	srv.approveAd(t, id)
	srv.assertPubliclyVisible(t, id)

	// Still reachable before the ad's lifetime has elapsed.
	beforeExpireCount, err := srv.expireOutdatedAds()
	require.NoError(t, err)
	assert.Equal(t, 0, beforeExpireCount)

	getBeforeResp := srv.get(t, "/classified-ads/"+id)
	defer getBeforeResp.Body.Close()
	assert.Equal(t, http.StatusOK, getBeforeResp.StatusCode)

	// Advance past the ad's lifetime and run the expiration sweep.
	srv.clock.Advance(addomain.AdLifetime + time.Hour)

	expiredCount, err := srv.expireOutdatedAds()
	require.NoError(t, err)
	assert.Equal(t, 1, expiredCount)

	// The expired ad is publicly invisible again.
	srv.assertPubliclyInvisible(t, id)

	// The expiration reached the moderation history through the public bus.
	assert.Equal(t, []string{"submitted", "approved", "published", "expired"}, srv.historyActions(t, id))
	entries := srv.historyEntries(t, id)
	lastEntry := entries[len(entries)-1]
	assert.Equal(t, moddomain.HistoryActionExpired, lastEntry.Action())
	assert.True(t, lastEntry.OccurredAt().After(entries[0].OccurredAt()), "the expired entry should be dated after publication")
}

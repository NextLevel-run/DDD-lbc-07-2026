package httpadapter

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/bcrypt"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/clock"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/application/query"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// handlerTestEnv exposes the mux plus the repository and clock, so tests can
// drive moderation transitions (approve/publish/challenge) that have no
// public HTTP endpoint in this bounded context.
type handlerTestEnv struct {
	mux   *http.ServeMux
	repo  *inmemory.InMemoryClassifiedAdRepository
	clock *clock.FixedClock
}

func setupHandlerTest(t *testing.T) *handlerTestEnv {
	t.Helper()

	repo := inmemory.NewInMemoryClassifiedAdRepository()
	hasher := bcrypt.NewBcryptPasswordHasher()
	fixedClock := clock.NewFixedClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()

	submit := command.BuildSubmitClassifiedAdCommand(repo, hasher, fixedClock, bus)
	makeOffer := command.BuildMakeOfferCommand(repo, fixedClock, bus)
	deleteAd := command.BuildDeleteClassifiedAdCommand(repo, hasher, fixedClock, bus)
	editAd := command.BuildEditClassifiedAdCommand(repo, hasher, fixedClock, bus)
	search := query.BuildSearchClassifiedAdsQuery(repo)
	get := query.BuildGetClassifiedAdQuery(repo)

	handler := NewHandler(submit, makeOffer, deleteAd, editAd, search, get)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return &handlerTestEnv{
		mux:   mux,
		repo:  repo,
		clock: fixedClock,
	}
}

func submitTestAd(t *testing.T, mux *http.ServeMux, email, password string) string {
	t.Helper()

	body := SubmitClassifiedAdRequest{
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

	req := httptest.NewRequest(http.MethodPost, "/classified-ads", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp SubmitClassifiedAdResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.NotEmpty(t, resp.ID)

	return resp.ID
}

// publishTestAd drives a submitted ad through the moderation happy path
// (approve + publish), as the moderation approval chain would.
func publishTestAd(t *testing.T, env *handlerTestEnv, id string) {
	t.Helper()

	adID, err := uuid.Parse(id)
	require.NoError(t, err)
	ad, err := env.repo.FindByID(adID)
	require.NoError(t, err)
	require.NoError(t, ad.Approve())
	require.NoError(t, ad.Publish(env.clock.Now()))
	require.NoError(t, env.repo.Save(ad))
}

// challengeTestAd transitions a submitted ad to challenged, as the moderation
// challenge chain would.
func challengeTestAd(t *testing.T, env *handlerTestEnv, id string) {
	t.Helper()

	adID, err := uuid.Parse(id)
	require.NoError(t, err)
	ad, err := env.repo.FindByID(adID)
	require.NoError(t, err)
	require.NoError(t, ad.Challenge())
	require.NoError(t, env.repo.Save(ad))
}

func validEditRequest(email, password string) EditClassifiedAdRequest {
	return EditClassifiedAdRequest{
		Email:        email,
		Password:     password,
		Title:        "Vélo hollandais (prix corrigé)",
		Description:  "Très bon état, prix ajusté",
		PriceInCents: 12000,
		ImageURLs:    []string{"http://example.com/img1.jpg"},
		Category:     "consumer_goods",
		ZipCode:      "75001",
		CityName:     "Paris",
	}
}

func editTestAd(t *testing.T, mux *http.ServeMux, id string, body EditClassifiedAdRequest) *httptest.ResponseRecorder {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/classified-ads/"+id, bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestSubmitClassifiedAd_Success(t *testing.T) {
	env := setupHandlerTest(t)

	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")

	assert.NotEmpty(t, id)
}

func TestGetClassifiedAd_UnknownID_NotFound(t *testing.T) {
	env := setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/classified-ads/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()
	env.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Error)
}

func TestGetClassifiedAd_SubmittedAd_NotFound(t *testing.T) {
	// A freshly submitted ad awaits moderation and is not publicly visible.
	env := setupHandlerTest(t)

	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")

	req := httptest.NewRequest(http.MethodGet, "/classified-ads/"+id, nil)
	rec := httptest.NewRecorder()
	env.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetClassifiedAd_PublishedAd_Success(t *testing.T) {
	env := setupHandlerTest(t)

	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")
	publishTestAd(t, env, id)

	req := httptest.NewRequest(http.MethodGet, "/classified-ads/"+id, nil)
	rec := httptest.NewRecorder()
	env.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp ClassifiedAdViewResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, id, resp.ID)
}

func TestDeleteClassifiedAd_WrongCredentials_Forbidden(t *testing.T) {
	env := setupHandlerTest(t)

	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")
	publishTestAd(t, env, id)

	deleteBody := DeleteClassifiedAdRequest{
		Email:    "seller@example.com",
		Password: "wrong-password",
		Reason:   "sold",
	}
	jsonBody, err := json.Marshal(deleteBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/classified-ads/"+id, bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	env.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestMakeOffer_OnDeletedAd_Conflict(t *testing.T) {
	env := setupHandlerTest(t)

	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")
	publishTestAd(t, env, id)

	deleteBody := DeleteClassifiedAdRequest{
		Email:    "seller@example.com",
		Password: "supersecret",
		Reason:   "sold",
	}
	jsonBody, err := json.Marshal(deleteBody)
	require.NoError(t, err)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/classified-ads/"+id, bytes.NewBuffer(jsonBody))
	deleteRec := httptest.NewRecorder()
	env.mux.ServeHTTP(deleteRec, deleteReq)
	require.Equal(t, http.StatusNoContent, deleteRec.Code)

	offerBody := MakeOfferRequest{
		BuyerEmail:    "buyer@example.com",
		BuyerPseudo:   "buyer1",
		AmountInCents: 12000,
		Message:       "Intéressé !",
	}
	offerJSON, err := json.Marshal(offerBody)
	require.NoError(t, err)

	offerReq := httptest.NewRequest(http.MethodPost, "/classified-ads/"+id+"/offers", bytes.NewBuffer(offerJSON))
	offerRec := httptest.NewRecorder()
	env.mux.ServeHTTP(offerRec, offerReq)

	assert.Equal(t, http.StatusConflict, offerRec.Code)
}

func TestSearchClassifiedAds_Success(t *testing.T) {
	env := setupHandlerTest(t)

	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")
	publishTestAd(t, env, id)

	req := httptest.NewRequest(http.MethodGet, "/classified-ads?category=consumer_goods", nil)
	rec := httptest.NewRecorder()
	env.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp SearchClassifiedAdsResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Len(t, resp.Items, 1)
}

func TestSearchClassifiedAds_DoesNotListSubmittedAds(t *testing.T) {
	env := setupHandlerTest(t)

	submitTestAd(t, env.mux, "seller@example.com", "supersecret")

	req := httptest.NewRequest(http.MethodGet, "/classified-ads", nil)
	rec := httptest.NewRecorder()
	env.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp SearchClassifiedAdsResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Empty(t, resp.Items, "Expected a submitted ad not to appear in public listings")
}

func TestEditClassifiedAd_Success(t *testing.T) {
	env := setupHandlerTest(t)

	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")
	challengeTestAd(t, env, id)

	rec := editTestAd(t, env.mux, id, validEditRequest("seller@example.com", "supersecret"))

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// The ad content was replaced and the ad is re-submitted for moderation
	adID, err := uuid.Parse(id)
	require.NoError(t, err)
	stored, err := env.repo.FindByID(adID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusSubmitted, stored.Status())
	assert.Equal(t, "Vélo hollandais (prix corrigé)", stored.Title())
	assert.Equal(t, int64(12000), stored.Price().AmountInCents())
}

func TestEditClassifiedAd_NotChallenged_Conflict(t *testing.T) {
	env := setupHandlerTest(t)

	// The ad is still submitted (not challenged), so it cannot be edited
	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")

	rec := editTestAd(t, env.mux, id, validEditRequest("seller@example.com", "supersecret"))

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestEditClassifiedAd_WrongCredentials_Forbidden(t *testing.T) {
	env := setupHandlerTest(t)

	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")
	challengeTestAd(t, env, id)

	rec := editTestAd(t, env.mux, id, validEditRequest("seller@example.com", "wrong-password"))

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestEditClassifiedAd_UnknownAd_NotFound(t *testing.T) {
	env := setupHandlerTest(t)

	rec := editTestAd(t, env.mux, "00000000-0000-0000-0000-000000000000", validEditRequest("seller@example.com", "supersecret"))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEditClassifiedAd_ValidationError_BadRequest(t *testing.T) {
	env := setupHandlerTest(t)

	id := submitTestAd(t, env.mux, "seller@example.com", "supersecret")
	challengeTestAd(t, env, id)

	body := validEditRequest("seller@example.com", "supersecret")
	body.Title = ""

	rec := editTestAd(t, env.mux, id, body)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEditClassifiedAd_InvalidBody_BadRequest(t *testing.T) {
	env := setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodPut, "/classified-ads/some-id", bytes.NewBufferString("{not-json"))
	rec := httptest.NewRecorder()
	env.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

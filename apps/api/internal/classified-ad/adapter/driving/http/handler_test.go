package httpadapter

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/bcrypt"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/clock"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/application/query"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

func setupHandlerTest(t *testing.T) (*Handler, *http.ServeMux) {
	t.Helper()

	repo := inmemory.NewInMemoryClassifiedAdRepository()
	hasher := bcrypt.NewBcryptPasswordHasher()
	fixedClock := clock.NewFixedClock(time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	bus := eventbus.NewSyncInMemoryEventBus()

	submit := command.BuildSubmitClassifiedAdCommand(repo, hasher, fixedClock, bus)
	makeOffer := command.BuildMakeOfferCommand(repo, fixedClock, bus)
	deleteAd := command.BuildDeleteClassifiedAdCommand(repo, hasher, fixedClock, bus)
	search := query.BuildSearchClassifiedAdsQuery(repo)
	get := query.BuildGetClassifiedAdQuery(repo)

	handler := NewHandler(submit, makeOffer, deleteAd, search, get)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return handler, mux
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

func TestSubmitClassifiedAd_Success(t *testing.T) {
	_, mux := setupHandlerTest(t)

	id := submitTestAd(t, mux, "seller@example.com", "supersecret")

	assert.NotEmpty(t, id)
}

func TestGetClassifiedAd_UnknownID_NotFound(t *testing.T) {
	_, mux := setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/classified-ads/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Error)
}

func TestDeleteClassifiedAd_WrongCredentials_Forbidden(t *testing.T) {
	_, mux := setupHandlerTest(t)

	id := submitTestAd(t, mux, "seller@example.com", "supersecret")

	deleteBody := DeleteClassifiedAdRequest{
		Email:    "seller@example.com",
		Password: "wrong-password",
		Reason:   "sold",
	}
	jsonBody, err := json.Marshal(deleteBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/classified-ads/"+id, bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestMakeOffer_OnDeletedAd_Conflict(t *testing.T) {
	_, mux := setupHandlerTest(t)

	id := submitTestAd(t, mux, "seller@example.com", "supersecret")

	deleteBody := DeleteClassifiedAdRequest{
		Email:    "seller@example.com",
		Password: "supersecret",
		Reason:   "sold",
	}
	jsonBody, err := json.Marshal(deleteBody)
	require.NoError(t, err)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/classified-ads/"+id, bytes.NewBuffer(jsonBody))
	deleteRec := httptest.NewRecorder()
	mux.ServeHTTP(deleteRec, deleteReq)
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
	mux.ServeHTTP(offerRec, offerReq)

	assert.Equal(t, http.StatusConflict, offerRec.Code)
}

func TestSearchClassifiedAds_Success(t *testing.T) {
	_, mux := setupHandlerTest(t)

	submitTestAd(t, mux, "seller@example.com", "supersecret")

	req := httptest.NewRequest(http.MethodGet, "/classified-ads?category=consumer_goods", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp SearchClassifiedAdsResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Len(t, resp.Items, 1)
}

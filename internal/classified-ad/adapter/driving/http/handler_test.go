package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPostClassifiedAdCommand struct {
	returnId   string
	returnErr  error
	calledWith command.PostClassifiedAdCommandArgs
}

func (m *mockPostClassifiedAdCommand) execute(args command.PostClassifiedAdCommandArgs) (string, error) {
	m.calledWith = args
	return m.returnId, m.returnErr
}

func setupHandlerTest() (*Handler, *mockPostClassifiedAdCommand) {
	mockCmd := &mockPostClassifiedAdCommand{returnId: "test-id-123"}
	handler := NewHandler(mockCmd.execute)
	return handler, mockCmd
}

func TestPostClassifiedAd_Success(t *testing.T) {
	handler, mockCmd := setupHandlerTest()
	mockCmd.returnId = "generated-id-123"

	body := PostClassifiedAdRequest{
		SellerId:      "seller-123",
		Title:         "Vélo VTT",
		Description:   "Vélo en très bon état",
		PriceAmount:   15000,
		PriceCurrency: "EUR",
		Category:      "Vehicles",
		PhotoURLs:     []string{"https://example.com/photo1.jpg"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/classified-ads", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()

	handler.PostClassifiedAd(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response PostClassifiedAdResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "generated-id-123", response.ID)

	assert.Equal(t, "seller-123", mockCmd.calledWith.SellerId)
	assert.Equal(t, "Vélo VTT", mockCmd.calledWith.Title)
}

func TestPostClassifiedAd_ValidationError(t *testing.T) {
	handler, mockCmd := setupHandlerTest()
	mockCmd.returnErr = domain.ErrEmptyTitle

	body := PostClassifiedAdRequest{SellerId: "seller-123"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/classified-ads", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()

	handler.PostClassifiedAd(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&response)
	assert.Equal(t, domain.ErrEmptyTitle.Error(), response.Error)
}

func TestPostClassifiedAd_MethodNotAllowed(t *testing.T) {
	handler, _ := setupHandlerTest()

	req := httptest.NewRequest(http.MethodGet, "/classified-ads", nil)
	rec := httptest.NewRecorder()

	handler.PostClassifiedAd(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestPostClassifiedAd_InvalidJSON(t *testing.T) {
	handler, _ := setupHandlerTest()

	req := httptest.NewRequest(http.MethodPost, "/classified-ads", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	handler.PostClassifiedAd(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

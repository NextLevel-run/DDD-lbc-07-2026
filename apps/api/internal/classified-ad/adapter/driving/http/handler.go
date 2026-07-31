package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/application/query"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
)

// Handler holds all command and query dependencies for the ClassifiedAd HTTP API.
type Handler struct {
	submitClassifiedAd  command.SubmitClassifiedAdCommand
	makeOffer           command.MakeOfferCommand
	deleteClassifiedAd  command.DeleteClassifiedAdCommand
	editClassifiedAd    command.EditClassifiedAdCommand
	searchClassifiedAds query.SearchClassifiedAdsQuery
	getClassifiedAd     query.GetClassifiedAdQuery
}

// NewHandler creates a Handler with all dependencies injected.
func NewHandler(
	submit command.SubmitClassifiedAdCommand,
	makeOffer command.MakeOfferCommand,
	delete_ command.DeleteClassifiedAdCommand,
	edit command.EditClassifiedAdCommand,
	search query.SearchClassifiedAdsQuery,
	get query.GetClassifiedAdQuery,
) *Handler {
	return &Handler{
		submitClassifiedAd:  submit,
		makeOffer:           makeOffer,
		deleteClassifiedAd:  delete_,
		editClassifiedAd:    edit,
		searchClassifiedAds: search,
		getClassifiedAd:     get,
	}
}

// RegisterRoutes registers the ClassifiedAd HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /classified-ads", h.SubmitClassifiedAd)
	mux.HandleFunc("GET /classified-ads", h.SearchClassifiedAds)
	mux.HandleFunc("GET /classified-ads/{id}", h.GetClassifiedAd)
	mux.HandleFunc("POST /classified-ads/{id}/offers", h.MakeOffer)
	mux.HandleFunc("DELETE /classified-ads/{id}", h.DeleteClassifiedAd)
	mux.HandleFunc("PUT /classified-ads/{id}", h.EditClassifiedAd)
}

// SubmitClassifiedAd handles POST /classified-ads.
func (h *Handler) SubmitClassifiedAd(w http.ResponseWriter, r *http.Request) {
	var req SubmitClassifiedAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.submitClassifiedAd(command.SubmitClassifiedAdCommandArgs{
		Title:          req.Title,
		Description:    req.Description,
		PriceInCents:   req.PriceInCents,
		SellerEmail:    req.SellerEmail,
		SellerPseudo:   req.SellerPseudo,
		SellerPassword: req.SellerPassword,
		ImageURLs:      req.ImageURLs,
		Category:       req.Category,
		ZipCode:        req.ZipCode,
		CityName:       req.CityName,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondSuccess(w, SubmitClassifiedAdResponse{ID: id}, http.StatusCreated)
}

// SearchClassifiedAds handles GET /classified-ads.
func (h *Handler) SearchClassifiedAds(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var category *string
	if v := q.Get("category"); v != "" {
		category = &v
	}
	var zipCode *string
	if v := q.Get("zip"); v != "" {
		zipCode = &v
	}
	var cityName *string
	if v := q.Get("city"); v != "" {
		cityName = &v
	}
	var keywords *string
	if v := q.Get("q"); v != "" {
		keywords = &v
	}

	var minPrice *int64
	if v := q.Get("minPrice"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			h.respondError(w, "invalid 'minPrice' parameter", http.StatusBadRequest)
			return
		}
		minPrice = &parsed
	}

	var maxPrice *int64
	if v := q.Get("maxPrice"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			h.respondError(w, "invalid 'maxPrice' parameter", http.StatusBadRequest)
			return
		}
		maxPrice = &parsed
	}

	limit := 0
	if v := q.Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			h.respondError(w, "invalid 'limit' parameter", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	offset := 0
	if v := q.Get("offset"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			h.respondError(w, "invalid 'offset' parameter", http.StatusBadRequest)
			return
		}
		offset = parsed
	}

	views, err := h.searchClassifiedAds(query.SearchClassifiedAdsQueryArgs{
		Category:        category,
		ZipCode:         zipCode,
		CityName:        cityName,
		Keywords:        keywords,
		MinPriceInCents: minPrice,
		MaxPriceInCents: maxPrice,
		SortBy:          q.Get("sortBy"),
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	items := make([]ClassifiedAdListItemResponse, len(views))
	for i, v := range views {
		items[i] = ClassifiedAdListItemResponse{
			ID:             v.ID,
			Title:          v.Title,
			PriceInCents:   v.PriceInCents,
			Category:       v.Category,
			CityName:       v.CityName,
			ZipCode:        v.ZipCode,
			FirstImageURL:  v.FirstImageURL,
			SubmissionDate: v.SubmissionDate,
		}
	}

	h.respondSuccess(w, SearchClassifiedAdsResponse{Items: items}, http.StatusOK)
}

// GetClassifiedAd handles GET /classified-ads/{id}.
func (h *Handler) GetClassifiedAd(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	view, err := h.getClassifiedAd(id)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondSuccess(w, ClassifiedAdViewResponse{
		ID:             view.ID,
		Title:          view.Title,
		Description:    view.Description,
		PriceInCents:   view.PriceInCents,
		Category:       view.Category,
		SellerPseudo:   view.SellerPseudo,
		ImageURLs:      view.ImageURLs,
		ZipCode:        view.ZipCode,
		CityName:       view.CityName,
		SubmissionDate: view.SubmissionDate,
	}, http.StatusOK)
}

// MakeOffer handles POST /classified-ads/{id}/offers.
func (h *Handler) MakeOffer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req MakeOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.makeOffer(command.MakeOfferCommandArgs{
		AdID:          id,
		BuyerEmail:    req.BuyerEmail,
		BuyerPseudo:   req.BuyerPseudo,
		AmountInCents: req.AmountInCents,
		Message:       req.Message,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DeleteClassifiedAd handles DELETE /classified-ads/{id}.
func (h *Handler) DeleteClassifiedAd(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req DeleteClassifiedAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.deleteClassifiedAd(command.DeleteClassifiedAdCommandArgs{
		AdID:     id,
		Email:    req.Email,
		Password: req.Password,
		Reason:   req.Reason,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EditClassifiedAd handles PUT /classified-ads/{id}.
func (h *Handler) EditClassifiedAd(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req EditClassifiedAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.editClassifiedAd(command.EditClassifiedAdCommandArgs{
		AdID:         id,
		Email:        req.Email,
		Password:     req.Password,
		Title:        req.Title,
		Description:  req.Description,
		PriceInCents: req.PriceInCents,
		ImageURLs:    req.ImageURLs,
		Category:     req.Category,
		ZipCode:      req.ZipCode,
		CityName:     req.CityName,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleDomainError maps domain errors to HTTP status codes and writes the response.
func (h *Handler) handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrClassifiedAdNotFound):
		h.respondError(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrInvalidCredentials):
		h.respondError(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrAdNotAvailable):
		h.respondError(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrCannotEdit):
		h.respondError(w, err.Error(), http.StatusConflict)
	case isValidationError(err):
		h.respondError(w, err.Error(), http.StatusBadRequest)
	default:
		h.respondError(w, err.Error(), http.StatusInternalServerError)
	}
}

// isValidationError reports whether err is one of the domain validation errors
// that should be mapped to a 400 Bad Request response.
func isValidationError(err error) bool {
	validationErrors := []error{
		domain.ErrEmptyTitle,
		domain.ErrTitleTooLong,
		domain.ErrEmptyDescription,
		domain.ErrDescriptionTooLong,
		domain.ErrNegativePrice,
		domain.ErrInvalidEmail,
		domain.ErrEmptyPseudo,
		domain.ErrPseudoTooLong,
		domain.ErrPasswordTooShort,
		domain.ErrInvalidZipCode,
		domain.ErrEmptyCityName,
		domain.ErrInvalidCategory,
		domain.ErrInvalidDeleteReason,
		domain.ErrTooManyImages,
		domain.ErrEmptyImageURL,
		domain.ErrEmptyOfferMessage,
		domain.ErrOfferMessageTooLong,
		domain.ErrNegativeOfferAmount,
	}
	for _, ve := range validationErrors {
		if errors.Is(err, ve) {
			return true
		}
	}
	return false
}

func (h *Handler) respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (h *Handler) respondSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

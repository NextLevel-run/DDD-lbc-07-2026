package http

import (
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/domain"
	"encoding/json"
	"errors"
	"net/http"
)

// Handler holds all command and query dependencies
type Handler struct {
	postClassifiedAdCommand command.PostClassifiedAdCommand
}

func NewHandler(postClassifiedAdCommand command.PostClassifiedAdCommand) *Handler {
	return &Handler{
		postClassifiedAdCommand: postClassifiedAdCommand,
	}
}

// PostClassifiedAd handles POST /classified-ads
func (h *Handler) PostClassifiedAd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PostClassifiedAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.postClassifiedAdCommand(command.PostClassifiedAdCommandArgs{
		SellerId:      req.SellerId,
		Title:         req.Title,
		Description:   req.Description,
		PriceAmount:   req.PriceAmount,
		PriceCurrency: req.PriceCurrency,
		Category:      req.Category,
		PhotoURLs:     req.PhotoURLs,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondSuccess(w, PostClassifiedAdResponse{
		ID:      id,
		Message: "Classified ad posted successfully",
	}, http.StatusCreated)
}

func (h *Handler) handleDomainError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	errorMessage := err.Error()

	switch {
	case errors.Is(err, domain.ErrEmptySellerId),
		errors.Is(err, domain.ErrEmptyTitle),
		errors.Is(err, domain.ErrTitleTooLong),
		errors.Is(err, domain.ErrEmptyDescription),
		errors.Is(err, domain.ErrNegativeAmount),
		errors.Is(err, domain.ErrInvalidCurrency),
		errors.Is(err, domain.ErrInvalidCategory),
		errors.Is(err, domain.ErrEmptyPhotoURL):
		statusCode = http.StatusBadRequest
	}

	h.respondError(w, errorMessage, statusCode)
}

func (h *Handler) respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (h *Handler) respondSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"

	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/application/query"
	"ddd-second-hand-marketplace/internal/moderation/domain"
)

// Handler holds all command and query dependencies for the Moderation HTTP API.
type Handler struct {
	claimModerationTask     command.ClaimModerationTaskCommand
	acceptClassifiedAd      command.AcceptClassifiedAdCommand
	rejectClassifiedAd      command.RejectClassifiedAdCommand
	challengeClassifiedAd   command.ChallengeClassifiedAdCommand
	listModerationTasks     query.ListModerationTasksQuery
	getModerationTaskDetail query.GetModerationTaskDetailQuery
}

// NewHandler creates a Handler with all dependencies injected.
func NewHandler(
	claim command.ClaimModerationTaskCommand,
	accept command.AcceptClassifiedAdCommand,
	reject command.RejectClassifiedAdCommand,
	challenge command.ChallengeClassifiedAdCommand,
	list query.ListModerationTasksQuery,
	getDetail query.GetModerationTaskDetailQuery,
) *Handler {
	return &Handler{
		claimModerationTask:     claim,
		acceptClassifiedAd:      accept,
		rejectClassifiedAd:      reject,
		challengeClassifiedAd:   challenge,
		listModerationTasks:     list,
		getModerationTaskDetail: getDetail,
	}
}

// RegisterRoutes registers the Moderation HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /moderation/tasks", h.ListModerationTasks)
	mux.HandleFunc("GET /moderation/tasks/{id}", h.GetModerationTaskDetail)
	mux.HandleFunc("POST /moderation/tasks/{id}/claim", h.ClaimModerationTask)
	mux.HandleFunc("POST /moderation/tasks/{id}/accept", h.AcceptClassifiedAd)
	mux.HandleFunc("POST /moderation/tasks/{id}/reject", h.RejectClassifiedAd)
	mux.HandleFunc("POST /moderation/tasks/{id}/challenge", h.ChallengeClassifiedAd)
}

// ListModerationTasks handles GET /moderation/tasks.
func (h *Handler) ListModerationTasks(w http.ResponseWriter, r *http.Request) {
	items, err := h.listModerationTasks()
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	tasks := make([]ModerationTaskListItemResponse, len(items))
	for i, item := range items {
		tasks[i] = ModerationTaskListItemResponse{
			ID:                item.ID,
			ClassifiedAdTitle: item.ClassifiedAdTitle,
			CreatedAt:         item.CreatedAt,
			Status:            item.Status,
			ClaimedBy:         item.ClaimedBy,
		}
	}

	h.respondSuccess(w, ListModerationTasksResponse{Tasks: tasks}, http.StatusOK)
}

// GetModerationTaskDetail handles GET /moderation/tasks/{id}.
func (h *Handler) GetModerationTaskDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	view, err := h.getModerationTaskDetail(id)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	history := make([]HistoryEntryResponse, len(view.History))
	for i, entry := range view.History {
		history[i] = HistoryEntryResponse{
			OccurredAt:  entry.OccurredAt,
			Action:      entry.Action,
			ModeratorID: entry.ModeratorID,
			Reason:      entry.Reason,
			Snapshot:    newSnapshotResponse(entry.Snapshot),
		}
	}

	h.respondSuccess(w, ModerationTaskDetailResponse{
		ID:             view.ID,
		ClassifiedAdID: view.ClassifiedAdID,
		CreatedAt:      view.CreatedAt,
		Status:         view.Status,
		ClaimedBy:      view.ClaimedBy,
		ModeratorID:    view.ModeratorID,
		ClaimedAt:      view.ClaimedAt,
		History:        history,
		LastSnapshot:   newSnapshotResponse(view.LastSnapshot),
	}, http.StatusOK)
}

// ClaimModerationTask handles POST /moderation/tasks/{id}/claim.
func (h *Handler) ClaimModerationTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req ClaimModerationTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.claimModerationTask(command.ClaimModerationTaskCommandArgs{
		TaskID:      id,
		ModeratorID: req.ModeratorID,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AcceptClassifiedAd handles POST /moderation/tasks/{id}/accept.
func (h *Handler) AcceptClassifiedAd(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req AcceptClassifiedAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.acceptClassifiedAd(command.AcceptClassifiedAdCommandArgs{
		TaskID:      id,
		ModeratorID: req.ModeratorID,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RejectClassifiedAd handles POST /moderation/tasks/{id}/reject.
func (h *Handler) RejectClassifiedAd(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req RejectClassifiedAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.rejectClassifiedAd(command.RejectClassifiedAdCommandArgs{
		TaskID:      id,
		ModeratorID: req.ModeratorID,
		Reason:      req.Reason,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChallengeClassifiedAd handles POST /moderation/tasks/{id}/challenge.
func (h *Handler) ChallengeClassifiedAd(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req ChallengeClassifiedAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.challengeClassifiedAd(command.ChallengeClassifiedAdCommandArgs{
		TaskID:      id,
		ModeratorID: req.ModeratorID,
		Reason:      req.Reason,
	})
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// newSnapshotResponse maps a snapshot view to its response DTO, or nil.
func newSnapshotResponse(snapshot *query.ClassifiedAdSnapshotView) *ClassifiedAdSnapshotResponse {
	if snapshot == nil {
		return nil
	}
	return &ClassifiedAdSnapshotResponse{
		Title:        snapshot.Title,
		Description:  snapshot.Description,
		PriceInCents: snapshot.PriceInCents,
		ImageURLs:    snapshot.ImageURLs,
		Category:     snapshot.Category,
		ZipCode:      snapshot.ZipCode,
		CityName:     snapshot.CityName,
		SellerEmail:  snapshot.SellerEmail,
		SellerPseudo: snapshot.SellerPseudo,
	}
}

// handleDomainError maps domain errors to HTTP status codes and writes the response.
func (h *Handler) handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrTaskAlreadyClaimed):
		h.respondError(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrNotTaskOwner):
		h.respondError(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrModerationTaskNotFound),
		errors.Is(err, domain.ErrModeratorNotFound):
		h.respondError(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrInvalidRejectReason),
		errors.Is(err, domain.ErrInvalidChallengeReason):
		h.respondError(w, err.Error(), http.StatusBadRequest)
	default:
		h.respondError(w, err.Error(), http.StatusInternalServerError)
	}
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

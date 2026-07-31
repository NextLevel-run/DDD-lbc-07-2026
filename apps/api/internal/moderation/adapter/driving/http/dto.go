package httpadapter

import "time"

// Request DTOs

// ClaimModerationTaskRequest is the request body for POST /moderation/tasks/{id}/claim.
// The moderator ID travels in the body: no authentication is in scope, the
// moderator is considered already authenticated upstream.
type ClaimModerationTaskRequest struct {
	ModeratorID string `json:"moderatorId"`
}

// AcceptClassifiedAdRequest is the request body for POST /moderation/tasks/{id}/accept.
type AcceptClassifiedAdRequest struct {
	ModeratorID string `json:"moderatorId"`
}

// RejectClassifiedAdRequest is the request body for POST /moderation/tasks/{id}/reject.
type RejectClassifiedAdRequest struct {
	ModeratorID string `json:"moderatorId"`
	Reason      string `json:"reason"`
}

// ChallengeClassifiedAdRequest is the request body for POST /moderation/tasks/{id}/challenge.
type ChallengeClassifiedAdRequest struct {
	ModeratorID string `json:"moderatorId"`
	Reason      string `json:"reason"`
}

// Response DTOs

// ModerationTaskListItemResponse is a single row of the moderation queue.
type ModerationTaskListItemResponse struct {
	ID                string    `json:"id"`
	ClassifiedAdTitle string    `json:"classifiedAdTitle"`
	CreatedAt         time.Time `json:"createdAt"`
	Status            string    `json:"status"`
	ClaimedBy         string    `json:"claimedBy,omitempty"`
}

// ListModerationTasksResponse is the response body for GET /moderation/tasks.
type ListModerationTasksResponse struct {
	Tasks []ModerationTaskListItemResponse `json:"tasks"`
}

// ClassifiedAdSnapshotResponse is the ad content captured at submission or
// edition time, as recorded in the history.
type ClassifiedAdSnapshotResponse struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	PriceInCents int64    `json:"priceInCents"`
	ImageURLs    []string `json:"imageUrls"`
	Category     string   `json:"category"`
	ZipCode      string   `json:"zipCode"`
	CityName     string   `json:"cityName"`
	SellerEmail  string   `json:"sellerEmail"`
	SellerPseudo string   `json:"sellerPseudo"`
}

// HistoryEntryResponse is one entry of the ad's moderation history.
type HistoryEntryResponse struct {
	OccurredAt  time.Time                     `json:"occurredAt"`
	Action      string                        `json:"action"`
	ModeratorID *string                       `json:"moderatorId,omitempty"`
	Reason      *string                       `json:"reason,omitempty"`
	Snapshot    *ClassifiedAdSnapshotResponse `json:"snapshot,omitempty"`
}

// ModerationTaskDetailResponse is the response body for GET /moderation/tasks/{id}.
type ModerationTaskDetailResponse struct {
	ID             string                        `json:"id"`
	ClassifiedAdID string                        `json:"classifiedAdId"`
	CreatedAt      time.Time                     `json:"createdAt"`
	Status         string                        `json:"status"`
	ClaimedBy      string                        `json:"claimedBy,omitempty"`
	ModeratorID    string                        `json:"moderatorId,omitempty"`
	ClaimedAt      *time.Time                    `json:"claimedAt,omitempty"`
	History        []HistoryEntryResponse        `json:"history"`
	LastSnapshot   *ClassifiedAdSnapshotResponse `json:"lastSnapshot,omitempty"`
}

// ErrorResponse is the standard error response body.
type ErrorResponse struct {
	Error string `json:"error"`
}

package domain

import "errors"

var (
	ErrEmptyModeratorFullName      = errors.New("moderator full name must not be empty")
	ErrInvalidModeratorID          = errors.New("moderator id must not be nil")
	ErrEmptyClassifiedAdID         = errors.New("classified ad id must not be empty")
	ErrTaskAlreadyClaimed          = errors.New("moderation task is already claimed")
	ErrNotTaskOwner                = errors.New("moderation task is not claimed by this moderator")
	ErrModerationTaskNotFound      = errors.New("moderation task not found")
	ErrModeratorNotFound           = errors.New("moderator not found")
	ErrClassifiedAdHistoryNotFound = errors.New("classified ad history not found")
	ErrInvalidRejectReason         = errors.New("invalid reject reason")
	ErrInvalidChallengeReason      = errors.New("invalid challenge reason")
	ErrInvalidHistoryAction        = errors.New("invalid history action")
)

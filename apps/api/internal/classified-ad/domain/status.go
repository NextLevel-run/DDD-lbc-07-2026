package domain

// Status represents the lifecycle state of a ClassifiedAd.
//
// Full lifecycle:
//
//	submitted → approved → published → deleted | expired
//	submitted → rejected → deleted (automatic, reason "rejected")
//	submitted → challenged → (edit) → submitted (re-submission)
//
// A new ad starts in StatusSubmitted, awaiting moderation. Moderation either
// approves it (then it is immediately published), rejects it (then it is
// automatically deleted) or challenges it (the seller must edit it, which
// re-submits it). StatusDeleted and StatusExpired are terminal.
type Status string

const (
	StatusSubmitted  Status = "submitted"
	StatusApproved   Status = "approved"
	StatusChallenged Status = "challenged"
	StatusRejected   Status = "rejected"
	StatusPublished  Status = "published"
	StatusDeleted    Status = "deleted"
	StatusExpired    Status = "expired"
)

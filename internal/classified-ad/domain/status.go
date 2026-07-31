package domain

// Status represents the lifecycle state of a ClassifiedAd.
type Status string

const (
	StatusPublished Status = "published"
	StatusDeleted   Status = "deleted"
	StatusExpired   Status = "expired"
)

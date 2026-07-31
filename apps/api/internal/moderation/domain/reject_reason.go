package domain

// RejectReason represents why a moderator rejected a classified ad.
type RejectReason string

const (
	RejectReasonInappropriateContent RejectReason = "inappropriate_content"
	RejectReasonSuspectPrice         RejectReason = "suspect_price"
	RejectReasonWrongCategory        RejectReason = "wrong_category"
)

// NewRejectReason validates and builds a RejectReason from a raw string.
func NewRejectReason(s string) (RejectReason, error) {
	switch RejectReason(s) {
	case RejectReasonInappropriateContent, RejectReasonSuspectPrice, RejectReasonWrongCategory:
		return RejectReason(s), nil
	default:
		return "", ErrInvalidRejectReason
	}
}

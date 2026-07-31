package domain

// DeleteReason represents why a classified ad was deleted.
type DeleteReason string

const (
	DeleteReasonSold         DeleteReason = "sold"
	DeleteReasonNoMoreToSell DeleteReason = "no_more_to_sell"
	DeleteReasonEdit         DeleteReason = "edit"
	DeleteReasonRejected     DeleteReason = "rejected"
)

// NewDeleteReason validates and builds a DeleteReason from a raw string.
func NewDeleteReason(s string) (DeleteReason, error) {
	switch DeleteReason(s) {
	case DeleteReasonSold, DeleteReasonNoMoreToSell, DeleteReasonEdit, DeleteReasonRejected:
		return DeleteReason(s), nil
	default:
		return "", ErrInvalidDeleteReason
	}
}

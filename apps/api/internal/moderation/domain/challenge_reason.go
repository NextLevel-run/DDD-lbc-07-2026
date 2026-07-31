package domain

// ChallengeReason represents why a moderator asked the seller to fix a classified ad.
type ChallengeReason string

const (
	ChallengeReasonPriceToVerify ChallengeReason = "price_to_verify"
	ChallengeReasonCategoryToFix ChallengeReason = "category_to_fix"
)

// NewChallengeReason validates and builds a ChallengeReason from a raw string.
func NewChallengeReason(s string) (ChallengeReason, error) {
	switch ChallengeReason(s) {
	case ChallengeReasonPriceToVerify, ChallengeReasonCategoryToFix:
		return ChallengeReason(s), nil
	default:
		return "", ErrInvalidChallengeReason
	}
}

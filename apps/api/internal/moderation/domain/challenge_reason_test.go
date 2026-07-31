package domain_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/moderation/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewChallengeReason(t *testing.T) {
	t.Run("valid reasons are accepted", func(t *testing.T) {
		for _, valid := range []string{"price_to_verify", "category_to_fix"} {
			r, err := domain.NewChallengeReason(valid)
			assert.NoError(t, err)
			assert.Equal(t, domain.ChallengeReason(valid), r)
		}
	})

	t.Run("invalid reason is rejected", func(t *testing.T) {
		_, err := domain.NewChallengeReason("just_because")
		assert.ErrorIs(t, err, domain.ErrInvalidChallengeReason)
	})

	t.Run("reject reason is not a valid challenge reason", func(t *testing.T) {
		_, err := domain.NewChallengeReason("suspect_price")
		assert.ErrorIs(t, err, domain.ErrInvalidChallengeReason)
	})

	t.Run("empty reason is rejected", func(t *testing.T) {
		_, err := domain.NewChallengeReason("")
		assert.ErrorIs(t, err, domain.ErrInvalidChallengeReason)
	})
}

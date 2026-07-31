package domain_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/moderation/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewRejectReason(t *testing.T) {
	t.Run("valid reasons are accepted", func(t *testing.T) {
		for _, valid := range []string{"inappropriate_content", "suspect_price", "wrong_category"} {
			r, err := domain.NewRejectReason(valid)
			assert.NoError(t, err)
			assert.Equal(t, domain.RejectReason(valid), r)
		}
	})

	t.Run("invalid reason is rejected", func(t *testing.T) {
		_, err := domain.NewRejectReason("too_ugly")
		assert.ErrorIs(t, err, domain.ErrInvalidRejectReason)
	})

	t.Run("challenge reason is not a valid reject reason", func(t *testing.T) {
		_, err := domain.NewRejectReason("price_to_verify")
		assert.ErrorIs(t, err, domain.ErrInvalidRejectReason)
	})

	t.Run("empty reason is rejected", func(t *testing.T) {
		_, err := domain.NewRejectReason("")
		assert.ErrorIs(t, err, domain.ErrInvalidRejectReason)
	})
}

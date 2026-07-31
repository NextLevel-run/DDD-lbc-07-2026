package domain_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewDeleteReason(t *testing.T) {
	t.Run("valid reasons are accepted", func(t *testing.T) {
		for _, valid := range []string{"sold", "no_more_to_sell", "edit"} {
			r, err := domain.NewDeleteReason(valid)
			assert.NoError(t, err)
			assert.Equal(t, domain.DeleteReason(valid), r)
		}
	})

	t.Run("invalid reason is rejected", func(t *testing.T) {
		_, err := domain.NewDeleteReason("changed_my_mind")
		assert.ErrorIs(t, err, domain.ErrInvalidDeleteReason)
	})

	t.Run("empty reason is rejected", func(t *testing.T) {
		_, err := domain.NewDeleteReason("")
		assert.ErrorIs(t, err, domain.ErrInvalidDeleteReason)
	})
}

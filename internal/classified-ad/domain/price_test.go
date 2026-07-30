package domain_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrice(t *testing.T) {
	t.Run("valid amount", func(t *testing.T) {
		p, err := domain.NewPrice(1000)
		require.NoError(t, err)
		assert.Equal(t, int64(1000), p.AmountInCents())
	})

	t.Run("zero is allowed", func(t *testing.T) {
		p, err := domain.NewPrice(0)
		require.NoError(t, err)
		assert.Equal(t, int64(0), p.AmountInCents())
	})

	t.Run("negative amount is rejected", func(t *testing.T) {
		_, err := domain.NewPrice(-1)
		assert.ErrorIs(t, err, domain.ErrNegativePrice)
	})
}

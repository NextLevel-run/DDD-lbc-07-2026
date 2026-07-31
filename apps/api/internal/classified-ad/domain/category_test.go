package domain_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewCategory(t *testing.T) {
	t.Run("valid categories are accepted", func(t *testing.T) {
		for _, valid := range []string{"immo", "auto", "consumer_goods", "holidays"} {
			c, err := domain.NewCategory(valid)
			assert.NoError(t, err)
			assert.Equal(t, domain.Category(valid), c)
		}
	})

	t.Run("invalid category is rejected", func(t *testing.T) {
		_, err := domain.NewCategory("furniture")
		assert.ErrorIs(t, err, domain.ErrInvalidCategory)
	})

	t.Run("empty category is rejected", func(t *testing.T) {
		_, err := domain.NewCategory("")
		assert.ErrorIs(t, err, domain.ErrInvalidCategory)
	})
}

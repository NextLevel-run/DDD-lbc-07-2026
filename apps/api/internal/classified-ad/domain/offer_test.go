package domain_test

import (
	"strings"
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
)

func TestValidateOfferMessage(t *testing.T) {
	t.Run("valid message", func(t *testing.T) {
		assert.NoError(t, domain.ValidateOfferMessage("I'd like to buy this."))
	})

	t.Run("empty message is rejected", func(t *testing.T) {
		assert.ErrorIs(t, domain.ValidateOfferMessage(""), domain.ErrEmptyOfferMessage)
	})

	t.Run("message exceeding 1000 characters is rejected", func(t *testing.T) {
		assert.ErrorIs(t, domain.ValidateOfferMessage(strings.Repeat("a", 1001)), domain.ErrOfferMessageTooLong)
	})

	t.Run("message of exactly 1000 characters is accepted", func(t *testing.T) {
		assert.NoError(t, domain.ValidateOfferMessage(strings.Repeat("a", 1000)))
	})
}

func TestValidateOfferAmount(t *testing.T) {
	t.Run("valid amount", func(t *testing.T) {
		assert.NoError(t, domain.ValidateOfferAmount(100))
	})

	t.Run("zero is allowed", func(t *testing.T) {
		assert.NoError(t, domain.ValidateOfferAmount(0))
	})

	t.Run("negative amount is rejected", func(t *testing.T) {
		assert.ErrorIs(t, domain.ValidateOfferAmount(-1), domain.ErrNegativeOfferAmount)
	})
}

package domain_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmail(t *testing.T) {
	t.Run("valid email is normalized to lowercase and trimmed", func(t *testing.T) {
		e, err := domain.NewEmail("  Foo.Bar@Example.COM  ")
		require.NoError(t, err)
		assert.Equal(t, "foo.bar@example.com", e.String())
	})

	t.Run("missing at sign is rejected", func(t *testing.T) {
		_, err := domain.NewEmail("not-an-email")
		assert.ErrorIs(t, err, domain.ErrInvalidEmail)
	})

	t.Run("missing domain dot is rejected", func(t *testing.T) {
		_, err := domain.NewEmail("foo@bar")
		assert.ErrorIs(t, err, domain.ErrInvalidEmail)
	})

	t.Run("empty string is rejected", func(t *testing.T) {
		_, err := domain.NewEmail("")
		assert.ErrorIs(t, err, domain.ErrInvalidEmail)
	})
}

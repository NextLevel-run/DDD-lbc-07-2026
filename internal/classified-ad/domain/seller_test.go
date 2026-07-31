package domain_test

import (
	"strings"
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validEmail(t *testing.T) domain.Email {
	t.Helper()
	e, err := domain.NewEmail("seller@example.com")
	require.NoError(t, err)
	return e
}

func validPassword(t *testing.T) domain.Password {
	t.Helper()
	p, err := domain.NewPassword("longenough", fakePasswordHasher{})
	require.NoError(t, err)
	return p
}

func TestNewSeller(t *testing.T) {
	email := validEmail(t)
	password := validPassword(t)

	t.Run("valid seller", func(t *testing.T) {
		s, err := domain.NewSeller(email, "seller-pseudo", password)
		require.NoError(t, err)
		assert.Equal(t, email, s.Email())
		assert.Equal(t, "seller-pseudo", s.Pseudo())
		assert.Equal(t, password, s.Password())
	})

	t.Run("empty pseudo is rejected", func(t *testing.T) {
		_, err := domain.NewSeller(email, "", password)
		assert.ErrorIs(t, err, domain.ErrEmptyPseudo)
	})

	t.Run("pseudo exceeding 30 characters is rejected", func(t *testing.T) {
		_, err := domain.NewSeller(email, strings.Repeat("a", 31), password)
		assert.ErrorIs(t, err, domain.ErrPseudoTooLong)
	})

	t.Run("pseudo of exactly 30 characters is accepted", func(t *testing.T) {
		_, err := domain.NewSeller(email, strings.Repeat("a", 30), password)
		assert.NoError(t, err)
	})
}

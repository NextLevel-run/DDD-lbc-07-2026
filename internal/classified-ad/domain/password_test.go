package domain_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPassword(t *testing.T) {
	hasher := fakePasswordHasher{}

	t.Run("valid password is hashed", func(t *testing.T) {
		p, err := domain.NewPassword("longenough", hasher)
		require.NoError(t, err)
		assert.True(t, p.Matches("longenough", hasher))
	})

	t.Run("too short password is rejected", func(t *testing.T) {
		_, err := domain.NewPassword("short", hasher)
		assert.ErrorIs(t, err, domain.ErrPasswordTooShort)
	})

	t.Run("exactly 8 characters is accepted", func(t *testing.T) {
		_, err := domain.NewPassword("12345678", hasher)
		assert.NoError(t, err)
	})
}

func TestPassword_Matches(t *testing.T) {
	hasher := fakePasswordHasher{}
	p, err := domain.NewPassword("correcthorse", hasher)
	require.NoError(t, err)

	t.Run("correct plaintext matches", func(t *testing.T) {
		assert.True(t, p.Matches("correcthorse", hasher))
	})

	t.Run("incorrect plaintext does not match", func(t *testing.T) {
		assert.False(t, p.Matches("wrongpassword", hasher))
	})
}

func TestPassword_DoesNotExposeHash(t *testing.T) {
	// Password must never expose its hash via String()/Stringer.
	// Compile-time check: domain.Password must not implement fmt.Stringer.
	var p domain.Password
	_, isStringer := any(p).(interface{ String() string })
	assert.False(t, isStringer, "domain.Password must not implement String()/Stringer")
}

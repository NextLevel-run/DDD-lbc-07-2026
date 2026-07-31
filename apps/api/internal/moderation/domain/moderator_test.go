package domain_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/moderation/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModerator(t *testing.T) {
	t.Run("valid moderator is created with a unique id", func(t *testing.T) {
		moderator, err := domain.NewModerator("Jane Doe")
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, moderator.ID())
		assert.Equal(t, "Jane Doe", moderator.FullName())
	})

	t.Run("two moderators get distinct ids", func(t *testing.T) {
		first, err := domain.NewModerator("Jane Doe")
		require.NoError(t, err)
		second, err := domain.NewModerator("John Smith")
		require.NoError(t, err)
		assert.NotEqual(t, first.ID(), second.ID())
	})

	t.Run("empty full name is rejected", func(t *testing.T) {
		_, err := domain.NewModerator("")
		assert.ErrorIs(t, err, domain.ErrEmptyModeratorFullName)
	})
}

func TestRehydrateModerator(t *testing.T) {
	t.Run("rebuilds a moderator with the given id", func(t *testing.T) {
		id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		moderator, err := domain.RehydrateModerator(id, "Jane Doe")
		require.NoError(t, err)
		assert.Equal(t, id, moderator.ID())
		assert.Equal(t, "Jane Doe", moderator.FullName())
	})

	t.Run("nil id is rejected", func(t *testing.T) {
		_, err := domain.RehydrateModerator(uuid.Nil, "Jane Doe")
		assert.ErrorIs(t, err, domain.ErrInvalidModeratorID)
	})

	t.Run("empty full name is rejected", func(t *testing.T) {
		id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		_, err := domain.RehydrateModerator(id, "")
		assert.ErrorIs(t, err, domain.ErrEmptyModeratorFullName)
	})
}

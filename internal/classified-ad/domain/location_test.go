package domain_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocation(t *testing.T) {
	t.Run("valid location", func(t *testing.T) {
		l, err := domain.NewLocation("75001", "Paris")
		require.NoError(t, err)
		assert.Equal(t, "75001", l.ZipCode())
		assert.Equal(t, "Paris", l.CityName())
	})

	t.Run("zip code with letters is rejected", func(t *testing.T) {
		_, err := domain.NewLocation("7500A", "Paris")
		assert.ErrorIs(t, err, domain.ErrInvalidZipCode)
	})

	t.Run("zip code too short is rejected", func(t *testing.T) {
		_, err := domain.NewLocation("7500", "Paris")
		assert.ErrorIs(t, err, domain.ErrInvalidZipCode)
	})

	t.Run("zip code too long is rejected", func(t *testing.T) {
		_, err := domain.NewLocation("750001", "Paris")
		assert.ErrorIs(t, err, domain.ErrInvalidZipCode)
	})

	t.Run("empty city name is rejected", func(t *testing.T) {
		_, err := domain.NewLocation("75001", "")
		assert.ErrorIs(t, err, domain.ErrEmptyCityName)
	})
}

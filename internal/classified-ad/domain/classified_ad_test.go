package domain_test

import (
	"strings"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validPlainPassword = "longenough"

func newValidSeller(t *testing.T) domain.Seller {
	t.Helper()
	email, err := domain.NewEmail("seller@example.com")
	require.NoError(t, err)
	password, err := domain.NewPassword(validPlainPassword, fakePasswordHasher{})
	require.NoError(t, err)
	seller, err := domain.NewSeller(email, "seller-pseudo", password)
	require.NoError(t, err)
	return seller
}

func newValidLocation(t *testing.T) domain.Location {
	t.Helper()
	loc, err := domain.NewLocation("75001", "Paris")
	require.NoError(t, err)
	return loc
}

func newValidClassifiedAd(t *testing.T, submittedAt time.Time) *domain.ClassifiedAd {
	t.Helper()
	ad, err := domain.NewClassifiedAd(
		"A great bike",
		"Barely used, excellent condition.",
		1500,
		newValidSeller(t),
		[]string{"http://example.com/img1.jpg"},
		domain.CategoryConsumerGoods,
		newValidLocation(t),
		domain.NewSubmissionDate(submittedAt),
	)
	require.NoError(t, err)
	return ad
}

func TestNewClassifiedAd(t *testing.T) {
	seller := newValidSeller(t)
	location := newValidLocation(t)
	now := time.Now()

	t.Run("valid ad is published immediately", func(t *testing.T) {
		ad, err := domain.NewClassifiedAd(
			"A great bike", "Barely used, excellent condition.", 1500,
			seller, []string{"http://example.com/img1.jpg"},
			domain.CategoryConsumerGoods, location, domain.NewSubmissionDate(now),
		)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, ad.ID())
		assert.Equal(t, domain.StatusPublished, ad.Status())
		assert.Equal(t, now, ad.PublishedAt())
		assert.True(t, ad.IsOnline())
		assert.Nil(t, ad.DeletedAt())
		assert.Nil(t, ad.ExpiredAt())
	})

	t.Run("empty title is rejected", func(t *testing.T) {
		_, err := domain.NewClassifiedAd(
			"", "description", 1500, seller, nil, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.ErrorIs(t, err, domain.ErrEmptyTitle)
	})

	t.Run("title exceeding 100 characters is rejected", func(t *testing.T) {
		_, err := domain.NewClassifiedAd(
			strings.Repeat("a", 101), "description", 1500, seller, nil, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.ErrorIs(t, err, domain.ErrTitleTooLong)
	})

	t.Run("title of exactly 100 characters is accepted", func(t *testing.T) {
		_, err := domain.NewClassifiedAd(
			strings.Repeat("a", 100), "description", 1500, seller, nil, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.NoError(t, err)
	})

	t.Run("empty description is rejected", func(t *testing.T) {
		_, err := domain.NewClassifiedAd(
			"title", "", 1500, seller, nil, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.ErrorIs(t, err, domain.ErrEmptyDescription)
	})

	t.Run("description exceeding 4000 characters is rejected", func(t *testing.T) {
		_, err := domain.NewClassifiedAd(
			"title", strings.Repeat("a", 4001), 1500, seller, nil, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.ErrorIs(t, err, domain.ErrDescriptionTooLong)
	})

	t.Run("description of exactly 4000 characters is accepted", func(t *testing.T) {
		_, err := domain.NewClassifiedAd(
			"title", strings.Repeat("a", 4000), 1500, seller, nil, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.NoError(t, err)
	})

	t.Run("negative price is rejected", func(t *testing.T) {
		_, err := domain.NewClassifiedAd(
			"title", "description", -1, seller, nil, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.ErrorIs(t, err, domain.ErrNegativePrice)
	})

	t.Run("more than 10 images is rejected", func(t *testing.T) {
		urls := make([]string, 11)
		for i := range urls {
			urls[i] = "http://example.com/img.jpg"
		}
		_, err := domain.NewClassifiedAd(
			"title", "description", 1500, seller, urls, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.ErrorIs(t, err, domain.ErrTooManyImages)
	})

	t.Run("exactly 10 images is accepted", func(t *testing.T) {
		urls := make([]string, 10)
		for i := range urls {
			urls[i] = "http://example.com/img.jpg"
		}
		_, err := domain.NewClassifiedAd(
			"title", "description", 1500, seller, urls, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.NoError(t, err)
	})

	t.Run("empty image url is rejected", func(t *testing.T) {
		_, err := domain.NewClassifiedAd(
			"title", "description", 1500, seller, []string{""}, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.ErrorIs(t, err, domain.ErrEmptyImageURL)
	})

	t.Run("no images is accepted", func(t *testing.T) {
		_, err := domain.NewClassifiedAd(
			"title", "description", 1500, seller, nil, domain.CategoryAuto, location, domain.NewSubmissionDate(now),
		)
		assert.NoError(t, err)
	})
}

func TestClassifiedAd_IsOnlineCanReceiveOfferIsExpirable(t *testing.T) {
	now := time.Now()
	ad := newValidClassifiedAd(t, now)

	t.Run("freshly published ad is online and can receive offers", func(t *testing.T) {
		assert.True(t, ad.IsOnline())
		assert.True(t, ad.CanReceiveOffer())
	})

	t.Run("is not expirable before the lifetime elapses", func(t *testing.T) {
		beforeThreshold := now.Add(domain.AdLifetime).Add(-time.Second)
		assert.False(t, ad.IsExpirable(beforeThreshold))
	})

	t.Run("is expirable exactly at the lifetime threshold", func(t *testing.T) {
		atThreshold := now.Add(domain.AdLifetime)
		assert.True(t, ad.IsExpirable(atThreshold))
	})

	t.Run("is expirable after the lifetime threshold", func(t *testing.T) {
		afterThreshold := now.Add(domain.AdLifetime).Add(time.Second)
		assert.True(t, ad.IsExpirable(afterThreshold))
	})

	t.Run("deleted ad is not online, cannot receive offers, and is not expirable", func(t *testing.T) {
		deletedAd := newValidClassifiedAd(t, now)
		ok, err := deletedAd.Delete(deletedAd.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, now)
		require.NoError(t, err)
		require.True(t, ok)

		assert.False(t, deletedAd.IsOnline())
		assert.False(t, deletedAd.CanReceiveOffer())
		assert.False(t, deletedAd.IsExpirable(now.Add(domain.AdLifetime)))
	})

	t.Run("expired ad is not online and cannot receive offers", func(t *testing.T) {
		expiredAd := newValidClassifiedAd(t, now)
		expiredAt := now.Add(domain.AdLifetime)
		require.True(t, expiredAd.Expire(expiredAt))

		assert.False(t, expiredAd.IsOnline())
		assert.False(t, expiredAd.CanReceiveOffer())
		assert.False(t, expiredAd.IsExpirable(expiredAt))
	})
}

func TestClassifiedAd_Expire(t *testing.T) {
	now := time.Now()

	t.Run("expiring before the threshold is a no-op", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		beforeThreshold := now.Add(domain.AdLifetime).Add(-time.Second)

		changed := ad.Expire(beforeThreshold)

		assert.False(t, changed)
		assert.Equal(t, domain.StatusPublished, ad.Status())
		assert.Nil(t, ad.ExpiredAt())
	})

	t.Run("expiring at or after the threshold transitions the ad", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		atThreshold := now.Add(domain.AdLifetime)

		changed := ad.Expire(atThreshold)

		assert.True(t, changed)
		assert.Equal(t, domain.StatusExpired, ad.Status())
		require.NotNil(t, ad.ExpiredAt())
		assert.Equal(t, atThreshold, *ad.ExpiredAt())
	})

	t.Run("expiring an already expired ad is a no-op", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		atThreshold := now.Add(domain.AdLifetime)
		require.True(t, ad.Expire(atThreshold))

		laterAttempt := atThreshold.Add(time.Hour)
		changed := ad.Expire(laterAttempt)

		assert.False(t, changed)
		assert.Equal(t, atThreshold, *ad.ExpiredAt())
	})

	t.Run("expiring a deleted ad is a no-op", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, now)
		require.NoError(t, err)
		require.True(t, ok)

		changed := ad.Expire(now.Add(domain.AdLifetime))

		assert.False(t, changed)
		assert.Equal(t, domain.StatusDeleted, ad.Status())
		assert.Nil(t, ad.ExpiredAt())
	})
}

func TestClassifiedAd_Delete(t *testing.T) {
	now := time.Now()

	t.Run("correct credentials delete the ad", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		deletedAt := now.Add(time.Hour)

		ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, deletedAt)

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, domain.StatusDeleted, ad.Status())
		require.NotNil(t, ad.DeletedAt())
		assert.Equal(t, deletedAt, *ad.DeletedAt())
		assert.Equal(t, domain.DeleteReasonSold, ad.DeleteReason())
	})

	t.Run("wrong email is rejected with no mutation", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		otherEmail, err := domain.NewEmail("someone-else@example.com")
		require.NoError(t, err)

		ok, err := ad.Delete(otherEmail, validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, now)

		assert.False(t, ok)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		assert.Equal(t, domain.StatusPublished, ad.Status())
		assert.Nil(t, ad.DeletedAt())
	})

	t.Run("wrong password is rejected with no mutation", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)

		ok, err := ad.Delete(ad.Seller().Email(), "wrong-password", domain.DeleteReasonSold, fakePasswordHasher{}, now)

		assert.False(t, ok)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		assert.Equal(t, domain.StatusPublished, ad.Status())
		assert.Nil(t, ad.DeletedAt())
	})

	t.Run("deleting twice is idempotent: second call returns false, nil and does not change state", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		firstDeleteAt := now.Add(time.Hour)

		ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, firstDeleteAt)
		require.NoError(t, err)
		require.True(t, ok)

		secondDeleteAt := now.Add(2 * time.Hour)
		ok, err = ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonEdit, fakePasswordHasher{}, secondDeleteAt)

		assert.False(t, ok)
		assert.NoError(t, err)
		require.NotNil(t, ad.DeletedAt())
		assert.Equal(t, firstDeleteAt, *ad.DeletedAt(), "deletedAt must not change on the idempotent second call")
		assert.Equal(t, domain.DeleteReasonSold, ad.DeleteReason(), "deleteReason must not change on the idempotent second call")
	})

	t.Run("deleting an already deleted ad does not check credentials", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, now)
		require.NoError(t, err)
		require.True(t, ok)

		// Wrong credentials on an already-deleted ad should still be a no-op, not an error.
		ok, err = ad.Delete(ad.Seller().Email(), "totally-wrong", domain.DeleteReasonEdit, fakePasswordHasher{}, now)

		assert.False(t, ok)
		assert.NoError(t, err)
	})
}

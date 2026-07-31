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

// newValidClassifiedAd builds a freshly submitted ad (StatusSubmitted).
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

// newPublishedClassifiedAd builds an ad that went through the full moderation
// happy path (submitted → approved → published), published at publishedAt.
func newPublishedClassifiedAd(t *testing.T, publishedAt time.Time) *domain.ClassifiedAd {
	t.Helper()
	ad := newValidClassifiedAd(t, publishedAt)
	require.NoError(t, ad.Approve())
	require.NoError(t, ad.Publish(publishedAt))
	return ad
}

func TestNewClassifiedAd(t *testing.T) {
	seller := newValidSeller(t)
	location := newValidLocation(t)
	now := time.Now()

	t.Run("valid ad is submitted awaiting moderation", func(t *testing.T) {
		ad, err := domain.NewClassifiedAd(
			"A great bike", "Barely used, excellent condition.", 1500,
			seller, []string{"http://example.com/img1.jpg"},
			domain.CategoryConsumerGoods, location, domain.NewSubmissionDate(now),
		)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, ad.ID())
		assert.Equal(t, domain.StatusSubmitted, ad.Status())
		assert.True(t, ad.PublishedAt().IsZero(), "publishedAt must not be set at creation")
		assert.False(t, ad.IsOnline())
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

func TestClassifiedAd_Approve(t *testing.T) {
	now := time.Now()

	t.Run("submitted ad is approved", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)

		err := ad.Approve()

		require.NoError(t, err)
		assert.Equal(t, domain.StatusApproved, ad.Status())
		assert.False(t, ad.IsOnline(), "an approved ad is not yet online")
		assert.True(t, ad.PublishedAt().IsZero(), "approval must not set publishedAt")
	})

	t.Run("approving a non-submitted ad is rejected with no mutation", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)

		err := ad.Approve()

		assert.ErrorIs(t, err, domain.ErrCannotApprove)
		assert.Equal(t, domain.StatusPublished, ad.Status())
	})

	t.Run("approving twice is rejected", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		require.NoError(t, ad.Approve())

		err := ad.Approve()

		assert.ErrorIs(t, err, domain.ErrCannotApprove)
		assert.Equal(t, domain.StatusApproved, ad.Status())
	})
}

func TestClassifiedAd_Publish(t *testing.T) {
	now := time.Now()

	t.Run("approved ad is published and publishedAt is set", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		require.NoError(t, ad.Approve())
		publishedAt := now.Add(time.Minute)

		err := ad.Publish(publishedAt)

		require.NoError(t, err)
		assert.Equal(t, domain.StatusPublished, ad.Status())
		assert.Equal(t, publishedAt, ad.PublishedAt())
		assert.True(t, ad.IsOnline())
	})

	t.Run("publishing a submitted ad is rejected with no mutation", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)

		err := ad.Publish(now)

		assert.ErrorIs(t, err, domain.ErrCannotPublish)
		assert.Equal(t, domain.StatusSubmitted, ad.Status())
		assert.True(t, ad.PublishedAt().IsZero())
		assert.False(t, ad.IsOnline())
	})

	t.Run("publishing twice is rejected", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)

		err := ad.Publish(now.Add(time.Hour))

		assert.ErrorIs(t, err, domain.ErrCannotPublish)
		assert.Equal(t, now, ad.PublishedAt(), "publishedAt must not change on a failed publish")
	})
}

func TestClassifiedAd_Reject(t *testing.T) {
	now := time.Now()

	t.Run("submitted ad is rejected", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)

		err := ad.Reject()

		require.NoError(t, err)
		assert.Equal(t, domain.StatusRejected, ad.Status())
		assert.False(t, ad.IsOnline())
	})

	t.Run("rejecting a non-submitted ad is rejected with no mutation", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		require.NoError(t, ad.Approve())

		err := ad.Reject()

		assert.ErrorIs(t, err, domain.ErrCannotReject)
		assert.Equal(t, domain.StatusApproved, ad.Status())
	})
}

func TestClassifiedAd_Challenge(t *testing.T) {
	now := time.Now()

	t.Run("submitted ad is challenged", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)

		err := ad.Challenge()

		require.NoError(t, err)
		assert.Equal(t, domain.StatusChallenged, ad.Status())
		assert.False(t, ad.IsOnline())
	})

	t.Run("challenging a non-submitted ad is rejected with no mutation", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)

		err := ad.Challenge()

		assert.ErrorIs(t, err, domain.ErrCannotChallenge)
		assert.Equal(t, domain.StatusPublished, ad.Status())
	})
}

func TestClassifiedAd_Edit(t *testing.T) {
	now := time.Now()

	newChallengedAd := func(t *testing.T) *domain.ClassifiedAd {
		t.Helper()
		ad := newValidClassifiedAd(t, now)
		require.NoError(t, ad.Challenge())
		return ad
	}

	t.Run("challenged ad is edited and re-submitted", func(t *testing.T) {
		ad := newChallengedAd(t)
		newLocation, err := domain.NewLocation("69001", "Lyon")
		require.NoError(t, err)

		err = ad.Edit(
			"A fixed bike",
			"Now with a fair price.",
			1200,
			[]string{"http://example.com/img2.jpg"},
			domain.CategoryAuto,
			newLocation,
		)

		require.NoError(t, err)
		assert.Equal(t, domain.StatusSubmitted, ad.Status())
		assert.Equal(t, "A fixed bike", ad.Title())
		assert.Equal(t, "Now with a fair price.", ad.Description())
		assert.Equal(t, int64(1200), ad.Price().AmountInCents())
		assert.Equal(t, []string{"http://example.com/img2.jpg"}, ad.ImageURLs())
		assert.Equal(t, domain.CategoryAuto, ad.Category())
		assert.Equal(t, newLocation, ad.Location())
		assert.False(t, ad.IsOnline())
	})

	t.Run("editing a non-challenged ad is rejected with no mutation", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)

		err := ad.Edit("New title", "New description.", 1000, nil, domain.CategoryAuto, newValidLocation(t))

		assert.ErrorIs(t, err, domain.ErrCannotEdit)
		assert.Equal(t, domain.StatusSubmitted, ad.Status())
		assert.Equal(t, "A great bike", ad.Title())
	})

	t.Run("invalid content is rejected with no mutation", func(t *testing.T) {
		testCases := []struct {
			name        string
			title       string
			description string
			price       int64
			imageURLs   []string
			expectedErr error
		}{
			{"empty title", "", "description", 1000, nil, domain.ErrEmptyTitle},
			{"title too long", strings.Repeat("a", 101), "description", 1000, nil, domain.ErrTitleTooLong},
			{"empty description", "title", "", 1000, nil, domain.ErrEmptyDescription},
			{"description too long", "title", strings.Repeat("a", 4001), 1000, nil, domain.ErrDescriptionTooLong},
			{"negative price", "title", "description", -1, nil, domain.ErrNegativePrice},
			{"too many images", "title", "description", 1000, make11ImageURLs(), domain.ErrTooManyImages},
			{"empty image url", "title", "description", 1000, []string{""}, domain.ErrEmptyImageURL},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ad := newChallengedAd(t)

				err := ad.Edit(tc.title, tc.description, tc.price, tc.imageURLs, domain.CategoryAuto, newValidLocation(t))

				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Equal(t, domain.StatusChallenged, ad.Status(), "a failed edit must leave the ad challenged")
				assert.Equal(t, "A great bike", ad.Title(), "a failed edit must not mutate the ad")
				assert.Equal(t, int64(1500), ad.Price().AmountInCents())
			})
		}
	})

	t.Run("edited ad can be challenged and edited again", func(t *testing.T) {
		ad := newChallengedAd(t)
		require.NoError(t, ad.Edit("First fix", "First correction.", 1000, nil, domain.CategoryAuto, newValidLocation(t)))
		require.NoError(t, ad.Challenge())

		err := ad.Edit("Second fix", "Second correction.", 900, nil, domain.CategoryAuto, newValidLocation(t))

		require.NoError(t, err)
		assert.Equal(t, domain.StatusSubmitted, ad.Status())
		assert.Equal(t, "Second fix", ad.Title())
	})
}

func make11ImageURLs() []string {
	urls := make([]string, 11)
	for i := range urls {
		urls[i] = "http://example.com/img.jpg"
	}
	return urls
}

func TestClassifiedAd_DeleteRejected(t *testing.T) {
	now := time.Now()

	t.Run("rejected ad is deleted with reason rejected", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		require.NoError(t, ad.Reject())
		deletedAt := now.Add(time.Minute)

		err := ad.DeleteRejected(deletedAt)

		require.NoError(t, err)
		assert.Equal(t, domain.StatusDeleted, ad.Status())
		require.NotNil(t, ad.DeletedAt())
		assert.Equal(t, deletedAt, *ad.DeletedAt())
		assert.Equal(t, domain.DeleteReasonRejected, ad.DeleteReason())
		assert.False(t, ad.IsOnline())
	})

	t.Run("deleting a non-rejected ad as rejected is refused with no mutation", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)

		err := ad.DeleteRejected(now)

		assert.ErrorIs(t, err, domain.ErrCannotDeleteRejected)
		assert.Equal(t, domain.StatusSubmitted, ad.Status())
		assert.Nil(t, ad.DeletedAt())
	})
}

func TestClassifiedAd_IsOnlineCanReceiveOfferIsExpirable(t *testing.T) {
	now := time.Now()
	ad := newPublishedClassifiedAd(t, now)

	t.Run("freshly published ad is online and can receive offers", func(t *testing.T) {
		assert.True(t, ad.IsOnline())
		assert.True(t, ad.CanReceiveOffer())
	})

	t.Run("submitted, approved, challenged and rejected ads are offline and cannot receive offers", func(t *testing.T) {
		submitted := newValidClassifiedAd(t, now)
		assert.False(t, submitted.IsOnline())
		assert.False(t, submitted.CanReceiveOffer())

		approved := newValidClassifiedAd(t, now)
		require.NoError(t, approved.Approve())
		assert.False(t, approved.IsOnline())
		assert.False(t, approved.CanReceiveOffer())

		challenged := newValidClassifiedAd(t, now)
		require.NoError(t, challenged.Challenge())
		assert.False(t, challenged.IsOnline())
		assert.False(t, challenged.CanReceiveOffer())

		rejected := newValidClassifiedAd(t, now)
		require.NoError(t, rejected.Reject())
		assert.False(t, rejected.IsOnline())
		assert.False(t, rejected.CanReceiveOffer())
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

	t.Run("submitted ad is never expirable", func(t *testing.T) {
		submitted := newValidClassifiedAd(t, now)
		assert.False(t, submitted.IsExpirable(now.Add(domain.AdLifetime)))
	})

	t.Run("deleted ad is not online, cannot receive offers, and is not expirable", func(t *testing.T) {
		deletedAd := newPublishedClassifiedAd(t, now)
		ok, err := deletedAd.Delete(deletedAd.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, now)
		require.NoError(t, err)
		require.True(t, ok)

		assert.False(t, deletedAd.IsOnline())
		assert.False(t, deletedAd.CanReceiveOffer())
		assert.False(t, deletedAd.IsExpirable(now.Add(domain.AdLifetime)))
	})

	t.Run("expired ad is not online and cannot receive offers", func(t *testing.T) {
		expiredAd := newPublishedClassifiedAd(t, now)
		expiredAt := now.Add(domain.AdLifetime)
		require.True(t, expiredAd.Expire(expiredAt))

		assert.False(t, expiredAd.IsOnline())
		assert.False(t, expiredAd.CanReceiveOffer())
		assert.False(t, expiredAd.IsExpirable(expiredAt))
	})
}

func TestClassifiedAd_IsOnlineIsRecomputedOnEveryMutation(t *testing.T) {
	now := time.Now()

	t.Run("isOnline flips to true exactly when Publish succeeds", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		require.False(t, ad.IsOnline())
		require.NoError(t, ad.Approve())
		require.False(t, ad.IsOnline())

		require.NoError(t, ad.Publish(now))

		assert.True(t, ad.IsOnline(), "isOnline must be recomputed as true once Publish transitions the ad to Published")
	})

	t.Run("isOnline flips to false exactly when Delete succeeds", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)
		require.True(t, ad.IsOnline())

		ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, now)
		require.NoError(t, err)
		require.True(t, ok)

		assert.False(t, ad.IsOnline(), "isOnline must be recomputed as false once Delete transitions the ad to Deleted")
	})

	t.Run("isOnline flips to false exactly when Expire succeeds", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)
		require.True(t, ad.IsOnline())

		require.True(t, ad.Expire(now.Add(domain.AdLifetime)))

		assert.False(t, ad.IsOnline(), "isOnline must be recomputed as false once Expire transitions the ad to Expired")
	})
}

func TestClassifiedAd_Expire(t *testing.T) {
	now := time.Now()

	t.Run("expiring before the threshold is a no-op", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)
		beforeThreshold := now.Add(domain.AdLifetime).Add(-time.Second)

		changed := ad.Expire(beforeThreshold)

		assert.False(t, changed)
		assert.Equal(t, domain.StatusPublished, ad.Status())
		assert.Nil(t, ad.ExpiredAt())
		assert.True(t, ad.IsOnline())
	})

	t.Run("expiring at or after the threshold transitions the ad", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)
		atThreshold := now.Add(domain.AdLifetime)

		changed := ad.Expire(atThreshold)

		assert.True(t, changed)
		assert.Equal(t, domain.StatusExpired, ad.Status())
		require.NotNil(t, ad.ExpiredAt())
		assert.Equal(t, atThreshold, *ad.ExpiredAt())
	})

	t.Run("expiring an already expired ad is a no-op", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)
		atThreshold := now.Add(domain.AdLifetime)
		require.True(t, ad.Expire(atThreshold))

		laterAttempt := atThreshold.Add(time.Hour)
		changed := ad.Expire(laterAttempt)

		assert.False(t, changed)
		assert.Equal(t, atThreshold, *ad.ExpiredAt())
		assert.False(t, ad.IsOnline())
	})

	t.Run("expiring a deleted ad is a no-op", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)
		ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, now)
		require.NoError(t, err)
		require.True(t, ok)

		changed := ad.Expire(now.Add(domain.AdLifetime))

		assert.False(t, changed)
		assert.Equal(t, domain.StatusDeleted, ad.Status())
		assert.Nil(t, ad.ExpiredAt())
		assert.False(t, ad.IsOnline())
	})

	t.Run("expiring a submitted ad is a no-op", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)

		changed := ad.Expire(now.Add(domain.AdLifetime))

		assert.False(t, changed)
		assert.Equal(t, domain.StatusSubmitted, ad.Status())
		assert.Nil(t, ad.ExpiredAt())
	})
}

func TestClassifiedAd_Delete(t *testing.T) {
	now := time.Now()

	t.Run("correct credentials delete the ad", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)
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
		ad := newPublishedClassifiedAd(t, now)
		otherEmail, err := domain.NewEmail("someone-else@example.com")
		require.NoError(t, err)

		ok, err := ad.Delete(otherEmail, validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, now)

		assert.False(t, ok)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		assert.Equal(t, domain.StatusPublished, ad.Status())
		assert.Nil(t, ad.DeletedAt())
		assert.True(t, ad.IsOnline())
	})

	t.Run("wrong password is rejected with no mutation", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)

		ok, err := ad.Delete(ad.Seller().Email(), "wrong-password", domain.DeleteReasonSold, fakePasswordHasher{}, now)

		assert.False(t, ok)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		assert.Equal(t, domain.StatusPublished, ad.Status())
		assert.Nil(t, ad.DeletedAt())
		assert.True(t, ad.IsOnline())
	})

	t.Run("deleting twice is idempotent: second call returns false, nil and does not change state", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)
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
		ad := newPublishedClassifiedAd(t, now)
		ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, now)
		require.NoError(t, err)
		require.True(t, ok)

		// Wrong credentials on an already-deleted ad should still be a no-op, not an error.
		ok, err = ad.Delete(ad.Seller().Email(), "totally-wrong", domain.DeleteReasonEdit, fakePasswordHasher{}, now)

		assert.False(t, ok)
		assert.NoError(t, err)
	})

	t.Run("seller can delete a submitted ad", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)

		ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonNoMoreToSell, fakePasswordHasher{}, now)

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, domain.StatusDeleted, ad.Status())
	})
}

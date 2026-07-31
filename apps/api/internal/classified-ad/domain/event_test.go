package domain_test

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClassifiedAdSubmittedEventFromClassifiedAd(t *testing.T) {
	now := time.Now()
	ad := newValidClassifiedAd(t, now)

	event := domain.NewClassifiedAdSubmittedEventFromClassifiedAd(ad)

	assert.Equal(t, "ClassifiedAdSubmitted", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, ad.Title(), event.Title)
	assert.Equal(t, ad.Description(), event.Description)
	assert.Equal(t, ad.Price().AmountInCents(), event.PriceInCents)
	assert.Equal(t, ad.ImageURLs(), event.ImageURLs)
	assert.Equal(t, string(ad.Category()), event.Category)
	assert.Equal(t, ad.Location().ZipCode(), event.ZipCode)
	assert.Equal(t, ad.Location().CityName(), event.CityName)
	assert.Equal(t, ad.Seller().Email().String(), event.SellerEmail)
	assert.Equal(t, ad.Seller().Pseudo(), event.SellerPseudo)
	assert.Equal(t, ad.SubmissionDate().Time(), event.OccurredAt)
}

func TestNewClassifiedAdEditedEventFromClassifiedAd(t *testing.T) {
	now := time.Now()
	ad := newValidClassifiedAd(t, now)
	require.NoError(t, ad.Challenge())
	newLocation, err := domain.NewLocation("69001", "Lyon")
	require.NoError(t, err)
	require.NoError(t, ad.Edit("A fixed bike", "Now with a fair price.", 1200, []string{"http://example.com/img2.jpg"}, domain.CategoryAuto, newLocation))
	occurredAt := now.Add(time.Hour)

	event := domain.NewClassifiedAdEditedEventFromClassifiedAd(ad, occurredAt)

	assert.Equal(t, "ClassifiedAdEdited", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, "A fixed bike", event.Title)
	assert.Equal(t, "Now with a fair price.", event.Description)
	assert.Equal(t, int64(1200), event.PriceInCents)
	assert.Equal(t, []string{"http://example.com/img2.jpg"}, event.ImageURLs)
	assert.Equal(t, string(domain.CategoryAuto), event.Category)
	assert.Equal(t, "69001", event.ZipCode)
	assert.Equal(t, "Lyon", event.CityName)
	assert.Equal(t, ad.Seller().Email().String(), event.SellerEmail)
	assert.Equal(t, ad.Seller().Pseudo(), event.SellerPseudo)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

func TestNewClassifiedAdApprovedEventFromClassifiedAd(t *testing.T) {
	now := time.Now()
	ad := newValidClassifiedAd(t, now)
	require.NoError(t, ad.Approve())
	occurredAt := now.Add(time.Hour)

	event := domain.NewClassifiedAdApprovedEventFromClassifiedAd(ad, occurredAt)

	assert.Equal(t, "ClassifiedAdApproved", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

func TestNewClassifiedAdRejectedEventFromClassifiedAd(t *testing.T) {
	now := time.Now()
	ad := newValidClassifiedAd(t, now)
	require.NoError(t, ad.Reject())
	occurredAt := now.Add(time.Hour)

	event := domain.NewClassifiedAdRejectedEventFromClassifiedAd(ad, occurredAt)

	assert.Equal(t, "ClassifiedAdRejected", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

func TestNewClassifiedAdChallengedEventFromClassifiedAd(t *testing.T) {
	now := time.Now()
	ad := newValidClassifiedAd(t, now)
	require.NoError(t, ad.Challenge())
	occurredAt := now.Add(time.Hour)

	event := domain.NewClassifiedAdChallengedEventFromClassifiedAd(ad, occurredAt)

	assert.Equal(t, "ClassifiedAdChallenged", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

func TestNewClassifiedAdPublishedEventFromClassifiedAd(t *testing.T) {
	now := time.Now()
	ad := newPublishedClassifiedAd(t, now)

	event := domain.NewClassifiedAdPublishedEventFromClassifiedAd(ad)

	assert.Equal(t, "ClassifiedAdPublished", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, ad.Title(), event.Title)
	assert.Equal(t, string(ad.Category()), event.Category)
	assert.Equal(t, ad.Seller().Email().String(), event.SellerEmail)
	assert.Equal(t, ad.Seller().Pseudo(), event.SellerPseudo)
	assert.Equal(t, ad.PublishedAt(), event.PublishedAt)
}

func TestNewBuyerOfferMadeEvent(t *testing.T) {
	now := time.Now()
	ad := newPublishedClassifiedAd(t, now)
	occurredAt := now.Add(time.Hour)

	event := domain.NewBuyerOfferMadeEvent(ad, "buyer@example.com", "buyer-pseudo", 2500, "I'll take it", occurredAt)

	assert.Equal(t, "BuyerOfferMade", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, ad.Title(), event.AdTitle)
	assert.Equal(t, ad.Seller().Email().String(), event.SellerEmail)
	assert.Equal(t, "buyer@example.com", event.BuyerEmail)
	assert.Equal(t, "buyer-pseudo", event.BuyerPseudo)
	assert.Equal(t, int64(2500), event.Amount)
	assert.Equal(t, "I'll take it", event.Message)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

func TestNewClassifiedAdDeletedEventFromClassifiedAd(t *testing.T) {
	now := time.Now()

	t.Run("after a seller deletion", func(t *testing.T) {
		ad := newPublishedClassifiedAd(t, now)
		deletedAt := now.Add(time.Hour)
		ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, deletedAt)
		require.NoError(t, err)
		require.True(t, ok)

		event := domain.NewClassifiedAdDeletedEventFromClassifiedAd(ad)

		assert.Equal(t, "ClassifiedAdDeleted", event.EventType())
		assert.Equal(t, ad.ID().String(), event.AdID)
		assert.Equal(t, string(domain.DeleteReasonSold), event.Reason)
		assert.Equal(t, deletedAt, event.DeletedAt)
	})

	t.Run("after an automatic deletion following a rejection", func(t *testing.T) {
		ad := newValidClassifiedAd(t, now)
		require.NoError(t, ad.Reject())
		deletedAt := now.Add(time.Hour)
		require.NoError(t, ad.DeleteRejected(deletedAt))

		event := domain.NewClassifiedAdDeletedEventFromClassifiedAd(ad)

		assert.Equal(t, "ClassifiedAdDeleted", event.EventType())
		assert.Equal(t, ad.ID().String(), event.AdID)
		assert.Equal(t, string(domain.DeleteReasonRejected), event.Reason)
		assert.Equal(t, deletedAt, event.DeletedAt)
	})
}

func TestNewClassifiedAdExpiredEventFromClassifiedAd(t *testing.T) {
	now := time.Now()
	ad := newPublishedClassifiedAd(t, now)
	expiredAt := now.Add(domain.AdLifetime)
	require.True(t, ad.Expire(expiredAt))

	event := domain.NewClassifiedAdExpiredEventFromClassifiedAd(ad)

	assert.Equal(t, "ClassifiedAdExpired", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, ad.Seller().Email().String(), event.SellerEmail)
	assert.Equal(t, expiredAt, event.ExpiredAt)
}

package domain_test

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClassifiedAdPublishedEventFromClassifiedAd(t *testing.T) {
	now := time.Now()
	ad := newValidClassifiedAd(t, now)

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
	ad := newValidClassifiedAd(t, now)
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
	ad := newValidClassifiedAd(t, now)
	deletedAt := now.Add(time.Hour)
	ok, err := ad.Delete(ad.Seller().Email(), validPlainPassword, domain.DeleteReasonSold, fakePasswordHasher{}, deletedAt)
	require.NoError(t, err)
	require.True(t, ok)

	event := domain.NewClassifiedAdDeletedEventFromClassifiedAd(ad)

	assert.Equal(t, "ClassifiedAdDeleted", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, string(domain.DeleteReasonSold), event.Reason)
	assert.Equal(t, deletedAt, event.DeletedAt)
}

func TestNewClassifiedAdExpiredEventFromClassifiedAd(t *testing.T) {
	now := time.Now()
	ad := newValidClassifiedAd(t, now)
	expiredAt := now.Add(domain.AdLifetime)
	require.True(t, ad.Expire(expiredAt))

	event := domain.NewClassifiedAdExpiredEventFromClassifiedAd(ad)

	assert.Equal(t, "ClassifiedAdExpired", event.EventType())
	assert.Equal(t, ad.ID().String(), event.AdID)
	assert.Equal(t, ad.Seller().Email().String(), event.SellerEmail)
	assert.Equal(t, expiredAt, event.ExpiredAt)
}

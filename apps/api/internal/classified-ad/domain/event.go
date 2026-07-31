package domain

import "time"

// ClassifiedAdPublishedEvent is emitted when a new classified ad is published.
type ClassifiedAdPublishedEvent struct {
	AdID         string
	Title        string
	Category     string
	SellerEmail  string
	SellerPseudo string
	PublishedAt  time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdPublishedEvent) EventType() string {
	return "ClassifiedAdPublished"
}

// NewClassifiedAdPublishedEventFromClassifiedAd builds a ClassifiedAdPublishedEvent from a ClassifiedAd.
func NewClassifiedAdPublishedEventFromClassifiedAd(ad *ClassifiedAd) *ClassifiedAdPublishedEvent {
	return &ClassifiedAdPublishedEvent{
		AdID:         ad.ID().String(),
		Title:        ad.Title(),
		Category:     string(ad.Category()),
		SellerEmail:  ad.Seller().Email().String(),
		SellerPseudo: ad.Seller().Pseudo(),
		PublishedAt:  ad.PublishedAt(),
	}
}

// BuyerOfferMadeEvent is emitted when a buyer makes an offer on a classified ad.
type BuyerOfferMadeEvent struct {
	AdID        string
	AdTitle     string
	SellerEmail string
	BuyerEmail  string
	BuyerPseudo string
	Amount      int64
	Message     string
	OccurredAt  time.Time
}

// EventType returns the event type name.
func (e *BuyerOfferMadeEvent) EventType() string {
	return "BuyerOfferMade"
}

// NewBuyerOfferMadeEvent builds a BuyerOfferMadeEvent from a ClassifiedAd and offer details.
func NewBuyerOfferMadeEvent(ad *ClassifiedAd, buyerEmail, buyerPseudo string, amountInCents int64, message string, occurredAt time.Time) *BuyerOfferMadeEvent {
	return &BuyerOfferMadeEvent{
		AdID:        ad.ID().String(),
		AdTitle:     ad.Title(),
		SellerEmail: ad.Seller().Email().String(),
		BuyerEmail:  buyerEmail,
		BuyerPseudo: buyerPseudo,
		Amount:      amountInCents,
		Message:     message,
		OccurredAt:  occurredAt,
	}
}

// ClassifiedAdDeletedEvent is emitted when a classified ad is deleted.
type ClassifiedAdDeletedEvent struct {
	AdID      string
	Reason    string
	DeletedAt time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdDeletedEvent) EventType() string {
	return "ClassifiedAdDeleted"
}

// NewClassifiedAdDeletedEventFromClassifiedAd builds a ClassifiedAdDeletedEvent from a ClassifiedAd.
func NewClassifiedAdDeletedEventFromClassifiedAd(ad *ClassifiedAd) *ClassifiedAdDeletedEvent {
	var deletedAt time.Time
	if ad.DeletedAt() != nil {
		deletedAt = *ad.DeletedAt()
	}
	return &ClassifiedAdDeletedEvent{
		AdID:      ad.ID().String(),
		Reason:    string(ad.DeleteReason()),
		DeletedAt: deletedAt,
	}
}

// ClassifiedAdExpiredEvent is emitted when a classified ad expires.
type ClassifiedAdExpiredEvent struct {
	AdID        string
	SellerEmail string
	ExpiredAt   time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdExpiredEvent) EventType() string {
	return "ClassifiedAdExpired"
}

// NewClassifiedAdExpiredEventFromClassifiedAd builds a ClassifiedAdExpiredEvent from a ClassifiedAd.
func NewClassifiedAdExpiredEventFromClassifiedAd(ad *ClassifiedAd) *ClassifiedAdExpiredEvent {
	var expiredAt time.Time
	if ad.ExpiredAt() != nil {
		expiredAt = *ad.ExpiredAt()
	}
	return &ClassifiedAdExpiredEvent{
		AdID:        ad.ID().String(),
		SellerEmail: ad.Seller().Email().String(),
		ExpiredAt:   expiredAt,
	}
}

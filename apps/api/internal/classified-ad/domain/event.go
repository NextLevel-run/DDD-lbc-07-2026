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

// ClassifiedAdSubmittedEvent is emitted when a new classified ad is submitted
// for moderation. It carries the full ad payload so publishers can build the
// public integration event without any repository access.
type ClassifiedAdSubmittedEvent struct {
	AdID         string
	Title        string
	Description  string
	PriceInCents int64
	ImageURLs    []string
	Category     string
	ZipCode      string
	CityName     string
	SellerEmail  string
	SellerPseudo string
	OccurredAt   time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdSubmittedEvent) EventType() string {
	return "ClassifiedAdSubmitted"
}

// NewClassifiedAdSubmittedEventFromClassifiedAd builds a ClassifiedAdSubmittedEvent from a ClassifiedAd.
func NewClassifiedAdSubmittedEventFromClassifiedAd(ad *ClassifiedAd) *ClassifiedAdSubmittedEvent {
	return &ClassifiedAdSubmittedEvent{
		AdID:         ad.ID().String(),
		Title:        ad.Title(),
		Description:  ad.Description(),
		PriceInCents: ad.Price().AmountInCents(),
		ImageURLs:    ad.ImageURLs(),
		Category:     string(ad.Category()),
		ZipCode:      ad.Location().ZipCode(),
		CityName:     ad.Location().CityName(),
		SellerEmail:  ad.Seller().Email().String(),
		SellerPseudo: ad.Seller().Pseudo(),
		OccurredAt:   ad.SubmissionDate().Time(),
	}
}

// ClassifiedAdEditedEvent is emitted when a seller edits a challenged ad,
// re-submitting it for moderation. It carries the full ad payload so
// publishers can build the public integration event without any repository
// access.
type ClassifiedAdEditedEvent struct {
	AdID         string
	Title        string
	Description  string
	PriceInCents int64
	ImageURLs    []string
	Category     string
	ZipCode      string
	CityName     string
	SellerEmail  string
	SellerPseudo string
	OccurredAt   time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdEditedEvent) EventType() string {
	return "ClassifiedAdEdited"
}

// NewClassifiedAdEditedEventFromClassifiedAd builds a ClassifiedAdEditedEvent from a ClassifiedAd.
func NewClassifiedAdEditedEventFromClassifiedAd(ad *ClassifiedAd, occurredAt time.Time) *ClassifiedAdEditedEvent {
	return &ClassifiedAdEditedEvent{
		AdID:         ad.ID().String(),
		Title:        ad.Title(),
		Description:  ad.Description(),
		PriceInCents: ad.Price().AmountInCents(),
		ImageURLs:    ad.ImageURLs(),
		Category:     string(ad.Category()),
		ZipCode:      ad.Location().ZipCode(),
		CityName:     ad.Location().CityName(),
		SellerEmail:  ad.Seller().Email().String(),
		SellerPseudo: ad.Seller().Pseudo(),
		OccurredAt:   occurredAt,
	}
}

// ClassifiedAdApprovedEvent is emitted when a classified ad is approved by
// moderation (submitted → approved).
type ClassifiedAdApprovedEvent struct {
	AdID       string
	OccurredAt time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdApprovedEvent) EventType() string {
	return "ClassifiedAdApproved"
}

// NewClassifiedAdApprovedEventFromClassifiedAd builds a ClassifiedAdApprovedEvent from a ClassifiedAd.
func NewClassifiedAdApprovedEventFromClassifiedAd(ad *ClassifiedAd, occurredAt time.Time) *ClassifiedAdApprovedEvent {
	return &ClassifiedAdApprovedEvent{
		AdID:       ad.ID().String(),
		OccurredAt: occurredAt,
	}
}

// ClassifiedAdRejectedEvent is emitted when a classified ad is rejected by
// moderation (submitted → rejected).
type ClassifiedAdRejectedEvent struct {
	AdID       string
	OccurredAt time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdRejectedEvent) EventType() string {
	return "ClassifiedAdRejected"
}

// NewClassifiedAdRejectedEventFromClassifiedAd builds a ClassifiedAdRejectedEvent from a ClassifiedAd.
func NewClassifiedAdRejectedEventFromClassifiedAd(ad *ClassifiedAd, occurredAt time.Time) *ClassifiedAdRejectedEvent {
	return &ClassifiedAdRejectedEvent{
		AdID:       ad.ID().String(),
		OccurredAt: occurredAt,
	}
}

// ClassifiedAdChallengedEvent is emitted when a classified ad is challenged by
// moderation (submitted → challenged), asking the seller for corrections.
type ClassifiedAdChallengedEvent struct {
	AdID       string
	OccurredAt time.Time
}

// EventType returns the event type name.
func (e *ClassifiedAdChallengedEvent) EventType() string {
	return "ClassifiedAdChallenged"
}

// NewClassifiedAdChallengedEventFromClassifiedAd builds a ClassifiedAdChallengedEvent from a ClassifiedAd.
func NewClassifiedAdChallengedEventFromClassifiedAd(ad *ClassifiedAd, occurredAt time.Time) *ClassifiedAdChallengedEvent {
	return &ClassifiedAdChallengedEvent{
		AdID:       ad.ID().String(),
		OccurredAt: occurredAt,
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

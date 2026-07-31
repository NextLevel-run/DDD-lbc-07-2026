package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdLifetime is the duration a published classified ad remains online before becoming expirable.
const AdLifetime = 90 * 24 * time.Hour

// ClassifiedAd is the aggregate root representing a second-hand item listed for sale.
type ClassifiedAd struct {
	id             uuid.UUID
	title          string
	description    string
	price          Price
	seller         Seller
	status         Status
	isOnline       bool
	imageURLs      []string
	submissionDate SubmissionDate
	publishedAt    time.Time
	category       Category
	location       Location
	deletedAt      *time.Time
	deleteReason   DeleteReason
	expiredAt      *time.Time
}

// NewClassifiedAd validates and builds a new ClassifiedAd, published immediately.
func NewClassifiedAd(
	title string,
	description string,
	priceInCents int64,
	seller Seller,
	imageURLs []string,
	category Category,
	location Location,
	submissionDate SubmissionDate,
) (*ClassifiedAd, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if len(title) > 100 {
		return nil, ErrTitleTooLong
	}
	if description == "" {
		return nil, ErrEmptyDescription
	}
	if len(description) > 4000 {
		return nil, ErrDescriptionTooLong
	}
	price, err := NewPrice(priceInCents)
	if err != nil {
		return nil, err
	}
	if len(imageURLs) > 10 {
		return nil, ErrTooManyImages
	}
	for _, url := range imageURLs {
		if url == "" {
			return nil, ErrEmptyImageURL
		}
	}

	ad := &ClassifiedAd{
		id:             uuid.New(),
		title:          title,
		description:    description,
		price:          price,
		seller:         seller,
		imageURLs:      imageURLs,
		submissionDate: submissionDate,
		publishedAt:    submissionDate.Time(),
		category:       category,
		location:       location,
	}
	ad.setStatus(StatusPublished)
	return ad, nil
}

// ID returns the aggregate identifier.
func (a *ClassifiedAd) ID() uuid.UUID {
	return a.id
}

// Title returns the ad's title.
func (a *ClassifiedAd) Title() string {
	return a.title
}

// Description returns the ad's description.
func (a *ClassifiedAd) Description() string {
	return a.description
}

// Price returns the ad's price.
func (a *ClassifiedAd) Price() Price {
	return a.price
}

// Seller returns the ad's seller.
func (a *ClassifiedAd) Seller() Seller {
	return a.seller
}

// Status returns the ad's current status.
func (a *ClassifiedAd) Status() Status {
	return a.status
}

// ImageURLs returns the ad's image URLs.
func (a *ClassifiedAd) ImageURLs() []string {
	return a.imageURLs
}

// SubmissionDate returns the ad's submission date.
func (a *ClassifiedAd) SubmissionDate() SubmissionDate {
	return a.submissionDate
}

// PublishedAt returns when the ad was published.
func (a *ClassifiedAd) PublishedAt() time.Time {
	return a.publishedAt
}

// Category returns the ad's category.
func (a *ClassifiedAd) Category() Category {
	return a.category
}

// Location returns the ad's location.
func (a *ClassifiedAd) Location() Location {
	return a.location
}

// DeletedAt returns when the ad was deleted, if applicable.
func (a *ClassifiedAd) DeletedAt() *time.Time {
	return a.deletedAt
}

// DeleteReason returns why the ad was deleted, if applicable.
func (a *ClassifiedAd) DeleteReason() DeleteReason {
	return a.deleteReason
}

// ExpiredAt returns when the ad expired, if applicable.
func (a *ClassifiedAd) ExpiredAt() *time.Time {
	return a.expiredAt
}

// IsOnline returns true if the ad is currently published and visible.
// It reflects the isOnline attribute, kept in sync with status by setStatus
// on every aggregate mutation.
func (a *ClassifiedAd) IsOnline() bool {
	return a.isOnline
}

// setStatus transitions the aggregate to newStatus, recomputing isOnline
// from it so the two fields can never drift apart.
func (a *ClassifiedAd) setStatus(newStatus Status) {
	a.status = newStatus
	a.isOnline = newStatus == StatusPublished
}

// CanReceiveOffer returns true if the ad can currently receive buyer offers.
func (a *ClassifiedAd) CanReceiveOffer() bool {
	return a.IsOnline()
}

// IsExpirable returns true if the ad is published and its lifetime has elapsed.
func (a *ClassifiedAd) IsExpirable(now time.Time) bool {
	return a.status == StatusPublished && !now.Before(a.publishedAt.Add(AdLifetime))
}

// Expire transitions the ad to StatusExpired if it is expirable. Returns false if not.
func (a *ClassifiedAd) Expire(now time.Time) bool {
	if !a.IsExpirable(now) {
		return false
	}
	a.setStatus(StatusExpired)
	a.expiredAt = &now
	return true
}

// Delete transitions the ad to StatusDeleted after verifying the seller's credentials.
// Returns false with no error if the ad is already deleted (idempotent no-op).
func (a *ClassifiedAd) Delete(email Email, password string, reason DeleteReason, hasher PasswordHasher, now time.Time) (bool, error) {
	if a.status == StatusDeleted {
		return false, nil
	}
	if email.String() != a.seller.Email().String() || !a.seller.Password().Matches(password, hasher) {
		return false, ErrInvalidCredentials
	}
	a.setStatus(StatusDeleted)
	a.deletedAt = &now
	a.deleteReason = reason
	return true, nil
}

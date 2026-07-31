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

// NewClassifiedAd validates and builds a new ClassifiedAd, submitted for moderation.
// The ad starts in StatusSubmitted with publishedAt unset — it is only set when
// the ad transitions from approved to published.
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
	price, err := validateContent(title, description, priceInCents, imageURLs)
	if err != nil {
		return nil, err
	}

	ad := &ClassifiedAd{
		id:             uuid.New(),
		title:          title,
		description:    description,
		price:          price,
		seller:         seller,
		imageURLs:      imageURLs,
		submissionDate: submissionDate,
		category:       category,
		location:       location,
	}
	ad.setStatus(StatusSubmitted)
	return ad, nil
}

// validateContent checks the editable content of a classified ad (title,
// description, price, images) and returns the validated Price. It is shared
// by NewClassifiedAd and Edit so both enforce the same business rules.
func validateContent(title, description string, priceInCents int64, imageURLs []string) (Price, error) {
	if title == "" {
		return Price{}, ErrEmptyTitle
	}
	if len(title) > 100 {
		return Price{}, ErrTitleTooLong
	}
	if description == "" {
		return Price{}, ErrEmptyDescription
	}
	if len(description) > 4000 {
		return Price{}, ErrDescriptionTooLong
	}
	price, err := NewPrice(priceInCents)
	if err != nil {
		return Price{}, err
	}
	if len(imageURLs) > 10 {
		return Price{}, ErrTooManyImages
	}
	for _, url := range imageURLs {
		if url == "" {
			return Price{}, ErrEmptyImageURL
		}
	}
	return price, nil
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

// PublishedAt returns when the ad was published. It is the zero time until
// the ad transitions from approved to published.
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

// Approve transitions the ad from StatusSubmitted to StatusApproved after a
// moderator accepted it. Returns ErrCannotApprove if the ad is not submitted.
func (a *ClassifiedAd) Approve() error {
	if a.status != StatusSubmitted {
		return ErrCannotApprove
	}
	a.setStatus(StatusApproved)
	return nil
}

// Publish transitions the ad from StatusApproved to StatusPublished, setting
// publishedAt to now. Returns ErrCannotPublish if the ad is not approved.
func (a *ClassifiedAd) Publish(now time.Time) error {
	if a.status != StatusApproved {
		return ErrCannotPublish
	}
	a.setStatus(StatusPublished)
	a.publishedAt = now
	return nil
}

// Reject transitions the ad from StatusSubmitted to StatusRejected after a
// moderator rejected it. Returns ErrCannotReject if the ad is not submitted.
// A rejected ad is then automatically deleted via DeleteRejected.
func (a *ClassifiedAd) Reject() error {
	if a.status != StatusSubmitted {
		return ErrCannotReject
	}
	a.setStatus(StatusRejected)
	return nil
}

// Challenge transitions the ad from StatusSubmitted to StatusChallenged when a
// moderator asks the seller for corrections. Returns ErrCannotChallenge if the
// ad is not submitted.
func (a *ClassifiedAd) Challenge() error {
	if a.status != StatusSubmitted {
		return ErrCannotChallenge
	}
	a.setStatus(StatusChallenged)
	return nil
}

// Edit replaces the editable content of a challenged ad (title, description,
// price, images, category, location) after validating it with the same rules
// as NewClassifiedAd, and re-submits the ad (StatusChallenged → StatusSubmitted).
// Returns ErrCannotEdit if the ad is not challenged; validation errors leave
// the ad unchanged.
func (a *ClassifiedAd) Edit(
	title string,
	description string,
	priceInCents int64,
	imageURLs []string,
	category Category,
	location Location,
) error {
	if a.status != StatusChallenged {
		return ErrCannotEdit
	}
	price, err := validateContent(title, description, priceInCents, imageURLs)
	if err != nil {
		return err
	}
	a.title = title
	a.description = description
	a.price = price
	a.imageURLs = imageURLs
	a.category = category
	a.location = location
	a.setStatus(StatusSubmitted)
	return nil
}

// DeleteRejected transitions the ad from StatusRejected to StatusDeleted with
// DeleteReasonRejected. It is the automatic system deletion following Reject —
// no seller credentials are involved. Returns ErrCannotDeleteRejected if the
// ad is not rejected.
func (a *ClassifiedAd) DeleteRejected(now time.Time) error {
	if a.status != StatusRejected {
		return ErrCannotDeleteRejected
	}
	a.setStatus(StatusDeleted)
	a.deletedAt = &now
	a.deleteReason = DeleteReasonRejected
	return nil
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

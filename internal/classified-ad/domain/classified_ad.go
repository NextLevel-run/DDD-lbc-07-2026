package domain

import (
	"errors"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ============================================
// ERRORS
// ============================================

var (
	ErrEmptyTitle          = errors.New("title cannot be empty")
	ErrEmptyDescription    = errors.New("description cannot be empty")
	ErrNegativePrice       = errors.New("price cannot be negative")
	ErrInvalidCategory     = errors.New("invalid category")
	ErrInvalidPhotoURL     = errors.New("invalid photo URL")
	ErrEmptyLocation       = errors.New("location cannot be empty")
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrEmptyNickname       = errors.New("nickname cannot be empty")
	ErrEmptyHashedPassword = errors.New("hashed password cannot be empty")
	// ErrEmptyPassword guards the seller's plaintext password before hashing.
	// It never enters the aggregate: the application checks it prior to hashing.
	ErrEmptyPassword = errors.New("password cannot be empty")
)

// ============================================
// AGGREGATE ROOT: ClassifiedAd
// ============================================

// ClassifiedAd is the aggregate root representing a second-hand ad posted by a
// seller. A seller has no account: all their contact details (email, nickname,
// hashed password) are carried by the ad itself through the SellerContact value
// object. A freshly posted ad always starts in PendingReview, awaiting moderation.
type ClassifiedAd struct {
	id          string
	version     int
	title       Title
	description Description
	price       Price
	category    Category
	photos      []Photo
	location    Location
	seller      SellerContact
	status      Status
	postedAt    time.Time
}

// NewClassifiedAd creates and validates a new ad. It takes primitive types and
// builds the value objects internally, keeping the public API simple and the
// validation encapsulated. The seller password MUST already be hashed: the
// plaintext must never enter the domain.
func NewClassifiedAd(
	title string,
	description string,
	priceInCents int64,
	category string,
	photoURLs []string,
	location string,
	sellerEmail string,
	sellerNickname string,
	sellerHashedPassword string,
) (*ClassifiedAd, error) {
	titleVO, err := NewTitle(title)
	if err != nil {
		return nil, err
	}
	descriptionVO, err := NewDescription(description)
	if err != nil {
		return nil, err
	}
	priceVO, err := NewPrice(priceInCents)
	if err != nil {
		return nil, err
	}
	categoryVO, err := NewCategory(category)
	if err != nil {
		return nil, err
	}
	photos := make([]Photo, 0, len(photoURLs))
	for _, rawURL := range photoURLs {
		photo, err := NewPhoto(rawURL)
		if err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}
	locationVO, err := NewLocation(location)
	if err != nil {
		return nil, err
	}
	seller, err := NewSellerContact(sellerEmail, sellerNickname, sellerHashedPassword)
	if err != nil {
		return nil, err
	}

	return &ClassifiedAd{
		id:          uuid.New().String(),
		version:     1,
		title:       titleVO,
		description: descriptionVO,
		price:       priceVO,
		category:    categoryVO,
		photos:      photos,
		location:    locationVO,
		seller:      seller,
		status:      StatusPendingReview,
		postedAt:    time.Now(),
	}, nil
}

func (a *ClassifiedAd) ID() string               { return a.id }
func (a *ClassifiedAd) Version() int             { return a.version }
func (a *ClassifiedAd) Title() Title             { return a.title }
func (a *ClassifiedAd) Description() Description { return a.description }
func (a *ClassifiedAd) Price() Price             { return a.price }
func (a *ClassifiedAd) Category() Category       { return a.category }
func (a *ClassifiedAd) Photos() []Photo          { return a.photos }
func (a *ClassifiedAd) Location() Location       { return a.location }
func (a *ClassifiedAd) Seller() SellerContact    { return a.seller }
func (a *ClassifiedAd) Status() Status           { return a.status }
func (a *ClassifiedAd) PostedAt() time.Time      { return a.postedAt }

// ============================================
// ENUM: Status
// ============================================

type Status string

const (
	StatusPendingReview Status = "PendingReview"
	StatusPublished     Status = "Published"
	StatusSold          Status = "Sold"
	StatusWithdrawn     Status = "Withdrawn"
)

// ============================================
// VALUE OBJECT: Title
// ============================================

type Title struct {
	value string
}

func NewTitle(value string) (Title, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Title{}, ErrEmptyTitle
	}
	return Title{value: trimmed}, nil
}

func (t Title) String() string { return t.value }

// ============================================
// VALUE OBJECT: Description
// ============================================

type Description struct {
	value string
}

func NewDescription(value string) (Description, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Description{}, ErrEmptyDescription
	}
	return Description{value: trimmed}, nil
}

func (d Description) String() string { return d.value }

// ============================================
// VALUE OBJECT: Price
// ============================================

// Price is a mono-currency amount stored in cents to avoid floating-point
// rounding issues. A free item is allowed (0), a negative price is not.
type Price struct {
	amountInCents int64
}

func NewPrice(amountInCents int64) (Price, error) {
	if amountInCents < 0 {
		return Price{}, ErrNegativePrice
	}
	return Price{amountInCents: amountInCents}, nil
}

func (p Price) AmountInCents() int64 { return p.amountInCents }

// ============================================
// VALUE OBJECT: Category (enum)
// ============================================

type Category string

const (
	CategoryElectronics Category = "Electronics"
	CategoryFurniture   Category = "Furniture"
	CategoryVehicles    Category = "Vehicles"
	CategoryClothing    Category = "Clothing"
	CategoryOther       Category = "Other"
)

var validCategories = map[Category]struct{}{
	CategoryElectronics: {},
	CategoryFurniture:   {},
	CategoryVehicles:    {},
	CategoryClothing:    {},
	CategoryOther:       {},
}

func NewCategory(value string) (Category, error) {
	category := Category(value)
	if _, exists := validCategories[category]; !exists {
		return "", ErrInvalidCategory
	}
	return category, nil
}

func (c Category) IsValid() bool {
	_, exists := validCategories[c]
	return exists
}

// ============================================
// VALUE OBJECT: Photo
// ============================================

type Photo struct {
	url string
}

func NewPhoto(rawURL string) (Photo, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return Photo{}, ErrInvalidPhotoURL
	}
	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return Photo{}, ErrInvalidPhotoURL
	}
	return Photo{url: trimmed}, nil
}

func (p Photo) URL() string { return p.url }

// ============================================
// VALUE OBJECT: Location
// ============================================

type Location struct {
	value string
}

func NewLocation(value string) (Location, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Location{}, ErrEmptyLocation
	}
	return Location{value: trimmed}, nil
}

func (l Location) String() string { return l.value }

// ============================================
// VALUE OBJECT: SellerContact
// ============================================

// SellerContact groups the seller's identity carried by the ad. The password is
// only ever held as a hash: HashedPassword cannot be built from plaintext, so
// the clear password can never accidentally be stored on the aggregate.
type SellerContact struct {
	email          Email
	nickname       Nickname
	hashedPassword HashedPassword
}

func NewSellerContact(email, nickname, hashedPassword string) (SellerContact, error) {
	emailVO, err := NewEmail(email)
	if err != nil {
		return SellerContact{}, err
	}
	nicknameVO, err := NewNickname(nickname)
	if err != nil {
		return SellerContact{}, err
	}
	hashedPasswordVO, err := NewHashedPassword(hashedPassword)
	if err != nil {
		return SellerContact{}, err
	}
	return SellerContact{
		email:          emailVO,
		nickname:       nicknameVO,
		hashedPassword: hashedPasswordVO,
	}, nil
}

func (s SellerContact) Email() Email                   { return s.email }
func (s SellerContact) Nickname() Nickname             { return s.nickname }
func (s SellerContact) HashedPassword() HashedPassword { return s.hashedPassword }

// ============================================
// VALUE OBJECT: Email
// ============================================

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: addr.Address}, nil
}

func (e Email) String() string { return e.value }

// ============================================
// VALUE OBJECT: Nickname
// ============================================

type Nickname struct {
	value string
}

func NewNickname(value string) (Nickname, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Nickname{}, ErrEmptyNickname
	}
	return Nickname{value: trimmed}, nil
}

func (n Nickname) String() string { return n.value }

// ============================================
// VALUE OBJECT: HashedPassword
// ============================================

// HashedPassword holds an already-hashed password. It deliberately offers no way
// to construct it from a plaintext value, keeping hashing an infrastructure
// concern (see the PasswordHasher port) and the domain free of clear passwords.
type HashedPassword struct {
	value string
}

func NewHashedPassword(hash string) (HashedPassword, error) {
	if strings.TrimSpace(hash) == "" {
		return HashedPassword{}, ErrEmptyHashedPassword
	}
	return HashedPassword{value: hash}, nil
}

func (h HashedPassword) String() string { return h.value }

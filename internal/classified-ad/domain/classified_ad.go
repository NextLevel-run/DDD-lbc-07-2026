package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ============================================
// ERRORS
// ============================================

var (
	ErrEmptySellerId        = errors.New("seller id cannot be empty")
	ErrEmptyTitle           = errors.New("title cannot be empty")
	ErrTitleTooLong         = errors.New("title cannot exceed 100 characters")
	ErrEmptyDescription     = errors.New("description cannot be empty")
	ErrNegativeAmount       = errors.New("price amount cannot be negative")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrInvalidCategory      = errors.New("invalid category")
	ErrEmptyPhotoURL        = errors.New("photo url cannot be empty")
	ErrClassifiedAdNotFound = errors.New("classified ad not found")
)

// ============================================
// AGGREGATE ROOT: ClassifiedAd
// ============================================

type ClassifiedAd struct {
	id          string
	version     int
	sellerId    string
	title       string
	description string
	price       Money
	category    Category
	photos      []Photo
	status      ClassifiedAdStatus
	postedAt    time.Time
}

func NewClassifiedAd(
	sellerId string,
	title string,
	description string,
	priceAmount int64,
	priceCurrency string,
	category string,
	photoURLs []string,
) (*ClassifiedAd, error) {
	if sellerId == "" {
		return nil, ErrEmptySellerId
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if len(title) > 100 {
		return nil, ErrTitleTooLong
	}
	if description == "" {
		return nil, ErrEmptyDescription
	}

	price, err := NewMoney(priceAmount, Currency(priceCurrency))
	if err != nil {
		return nil, err
	}

	cat, err := NewCategory(category)
	if err != nil {
		return nil, err
	}

	photos := make([]Photo, 0, len(photoURLs))
	for _, url := range photoURLs {
		photo, err := NewPhoto(url)
		if err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}

	return &ClassifiedAd{
		id:          uuid.New().String(),
		version:     1,
		sellerId:    sellerId,
		title:       title,
		description: description,
		price:       price,
		category:    cat,
		photos:      photos,
		status:      ClassifiedAdPublished,
		postedAt:    time.Now(),
	}, nil
}

func (c *ClassifiedAd) ID() string                 { return c.id }
func (c *ClassifiedAd) Version() int               { return c.version }
func (c *ClassifiedAd) SellerId() string           { return c.sellerId }
func (c *ClassifiedAd) Title() string              { return c.title }
func (c *ClassifiedAd) Description() string        { return c.description }
func (c *ClassifiedAd) Price() Money               { return c.price }
func (c *ClassifiedAd) Category() Category         { return c.category }
func (c *ClassifiedAd) Photos() []Photo            { return c.photos }
func (c *ClassifiedAd) Status() ClassifiedAdStatus { return c.status }
func (c *ClassifiedAd) PostedAt() time.Time        { return c.postedAt }

// ============================================
// ENUM: ClassifiedAdStatus
// ============================================

type ClassifiedAdStatus string

const (
	ClassifiedAdPublished ClassifiedAdStatus = "Published"
)

// ============================================
// VALUE OBJECT: Category
// ============================================

type Category string

const (
	CategoryVehicles    Category = "Vehicles"
	CategoryRealEstate  Category = "RealEstate"
	CategoryElectronics Category = "Electronics"
	CategoryFurniture   Category = "Furniture"
	CategoryFashion     Category = "Fashion"
	CategoryOther       Category = "Other"
)

var validCategories = map[Category]struct{}{
	CategoryVehicles:    {},
	CategoryRealEstate:  {},
	CategoryElectronics: {},
	CategoryFurniture:   {},
	CategoryFashion:     {},
	CategoryOther:       {},
}

func NewCategory(value string) (Category, error) {
	category := Category(value)
	if _, exists := validCategories[category]; !exists {
		return "", ErrInvalidCategory
	}
	return category, nil
}

// ============================================
// VALUE OBJECT: Money
// ============================================

type Money struct {
	amount   int64 // in cents
	currency Currency
}

func NewMoney(amount int64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	if !currency.IsValid() {
		return Money{}, ErrInvalidCurrency
	}
	return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() int64      { return m.amount }
func (m Money) Currency() Currency { return m.currency }

// ============================================
// VALUE OBJECT: Currency (enum)
// ============================================

type Currency string

const (
	EUR Currency = "EUR"
)

var validCurrencies = map[Currency]struct{}{
	EUR: {},
}

func (c Currency) IsValid() bool {
	_, exists := validCurrencies[c]
	return exists
}

// ============================================
// VALUE OBJECT: Photo
// ============================================

type Photo struct {
	url string
}

func NewPhoto(url string) (Photo, error) {
	if url == "" {
		return Photo{}, ErrEmptyPhotoURL
	}
	return Photo{url: url}, nil
}

func (p Photo) URL() string { return p.url }

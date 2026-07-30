package domain

import (
	"time"

	"github.com/google/uuid"
)

// PasswordHasher is a port for hashing and comparing passwords.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

// Clock is a port abstracting the current time.
type Clock interface {
	Now() time.Time
}

// SearchCriteria describes the filters, sorting and pagination applied to a Search query.
type SearchCriteria struct {
	Category        *Category
	ZipCode         *string
	CityName        *string
	MinPriceInCents *int64
	MaxPriceInCents *int64
	Keywords        *string // matches Title or Description, case-insensitive substring
	SortBy          string  // "date_desc" (default), "date_asc", "price_asc", "price_desc"
	Limit           int
	Offset          int
}

// ClassifiedAdRepository is the persistence port for ClassifiedAd aggregates.
type ClassifiedAdRepository interface {
	Save(ad *ClassifiedAd) error
	FindByID(id uuid.UUID) (*ClassifiedAd, error)
	FindExpirable(now time.Time) ([]*ClassifiedAd, error)
	Search(criteria SearchCriteria) ([]*ClassifiedAd, error)
}

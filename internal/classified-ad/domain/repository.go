package domain

// FindAllFilters contains optional filters when querying ads.
// A nil pointer means "no filter" on that field.
type FindAllFilters struct {
	Status   *Status
	Category *Category
}

// ClassifiedAdRepository is the persistence port for the ClassifiedAd aggregate.
// The domain defines WHAT it needs; driven adapters decide HOW to provide it.
type ClassifiedAdRepository interface {
	// Save persists an ad (create or update).
	Save(ad *ClassifiedAd) error

	// GetById retrieves a single ad by its unique identifier.
	GetById(id string) (*ClassifiedAd, error)

	// FindAll retrieves ads with optional filtering.
	FindAll(filters FindAllFilters) ([]*ClassifiedAd, error)
}

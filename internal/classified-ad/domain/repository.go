package domain

// ClassifiedAdRepository defines persistence operations for ClassifiedAd aggregate
type ClassifiedAdRepository interface {
	// Save persists a classified ad (create or update)
	Save(classifiedAd *ClassifiedAd) error

	// GetById retrieves a single classified ad by its unique identifier
	GetById(id string) (*ClassifiedAd, error)
}

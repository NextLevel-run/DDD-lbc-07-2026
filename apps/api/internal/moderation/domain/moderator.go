package domain

import "github.com/google/uuid"

// Moderator is the aggregate root representing a person reviewing classified ads.
// No authentication is handled in this scope — the moderator is considered
// already authenticated upstream; the fullName exists for traceability.
type Moderator struct {
	id       uuid.UUID
	fullName string
}

// NewModerator validates and builds a new Moderator.
func NewModerator(fullName string) (*Moderator, error) {
	if fullName == "" {
		return nil, ErrEmptyModeratorFullName
	}
	return &Moderator{
		id:       uuid.New(),
		fullName: fullName,
	}, nil
}

// RehydrateModerator rebuilds a Moderator with a known identity, e.g. when
// loading from persistence or seeding fixtures with fixed IDs.
func RehydrateModerator(id uuid.UUID, fullName string) (*Moderator, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidModeratorID
	}
	if fullName == "" {
		return nil, ErrEmptyModeratorFullName
	}
	return &Moderator{
		id:       id,
		fullName: fullName,
	}, nil
}

// ID returns the aggregate identifier.
func (m *Moderator) ID() uuid.UUID {
	return m.id
}

// FullName returns the moderator's full name.
func (m *Moderator) FullName() string {
	return m.fullName
}

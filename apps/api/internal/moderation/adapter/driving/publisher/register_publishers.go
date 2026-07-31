package publisher

import (
	"ddd-second-hand-marketplace/internal/shared"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// RegisterPublishers registers every Moderation → Public event publisher on the
// internal bus. This is the single wiring entry point for main.go: internal
// domain events never cross the context boundary, only the public DTOs
// published here do.
func RegisterPublishers(internalBus eventbus.Bus, publicBus shared.PublicEventBus) error {
	if err := NewClassifiedAdApprovedPublisher(internalBus, publicBus); err != nil {
		return err
	}
	if err := NewClassifiedAdRejectedPublisher(internalBus, publicBus); err != nil {
		return err
	}
	return NewClassifiedAdChallengedPublisher(internalBus, publicBus)
}

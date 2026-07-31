package shared

import "ddd-second-hand-marketplace/pkg/eventbus"

// PublicEventBus is the inter-bounded-context integration bus.
//
// It is a second, separate bus instance from each bounded context's internal
// event bus: internal domain events never cross a context boundary. Instead,
// each bounded context has a publisher adapter that consumes its own internal
// events and publishes the corresponding public event DTOs (defined in this
// package) on the public bus, where other bounded contexts' consumers pick
// them up.
//
// It reuses the generic pkg/eventbus.Bus contract, so any Bus implementation
// (sync or async in-memory) can serve as the public bus.
type PublicEventBus = eventbus.Bus

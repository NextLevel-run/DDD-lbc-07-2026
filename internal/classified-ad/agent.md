# 🍀 Classified Ad Domain

**Type:** Core Domain
**Description:** Sellers post classified ads for second-hand items on the marketplace.
**Status:** Partial — only "seller posts a classified ad" is implemented. Moderation/review, purchase flow, reporting, advertising and expiration are not implemented yet (see event-storming board for the full scope).

---

## Ubiquitous Language

### **Classified Ad**
The aggregate root representing an item a seller offers for sale. It carries a title, description, price, category, and photos. A classified ad is uniquely identified by an `id` (UUID) and tracks an optimistic-concurrency `version`. Once created it is immediately `Published` — there is no draft or moderation step in the current implementation.

### **Seller**
The user posting the classified ad, referenced by `sellerId` (a plain string identifier; no `Seller` entity exists in this bounded context — sellers are owned by another, not-yet-implemented bounded context).

### **Money**
A value object pairing an `amount` (int64, expressed in cents) with a `Currency`. Validates that the amount is not negative and that the currency is valid.

### **Currency**
An enum value object restricted to `EUR` for now. Any other value is rejected as invalid.

### **Category**
An enum value object classifying the item being sold. Valid values: `Vehicles`, `RealEstate`, `Electronics`, `Furniture`, `Fashion`, `Other`.

### **Photo**
A value object wrapping a photo URL attached to a classified ad. The URL must not be empty. A classified ad can have zero or more photos.

### **Classified Ad Status**
An enum tracking the lifecycle state of the ad. Currently only `Published` exists — the ad is published as soon as it is created. Moderation states (e.g. pending review, rejected) are part of the future event-storming board but not implemented.

---

## User Journeys

### Seller Posts a Classified Ad
1. Seller submits title, description, price (amount + currency), category, and optional photo URLs.
2. The system validates all fields and constructs a `ClassifiedAd` in `Published` status.
3. The ad is persisted via the `ClassifiedAdRepository`.
4. A `ClassifiedAdPostedEvent` is emitted after successful persistence.
5. The seller receives the newly created ad's `id`.

---

## Architecture

### Folder Structure

```
internal/classified-ad/
├── domain/
│   ├── classified_ad.go
│   ├── classified_ad_test.go
│   ├── event.go
│   └── repository.go
├── application/
│   ├── command/
│   │   ├── post_classified_ad.go
│   │   └── post_classified_ad_test.go
│   └── query/
│       └── .gitkeep            # no queries implemented yet
└── adapter/
    ├── driven/
    │   └── inmemory/
    │       └── repository.go
    └── driving/
        ├── http/
        │   ├── dto.go
        │   ├── handler.go
        │   └── handler_test.go
        └── consumer/
            └── .gitkeep         # no event consumers implemented yet
```

---

## Domain

### Aggregates

| Aggregate | Description | Validation Rules | Link |
|-----------|-------------|-------------------|------|
| `ClassifiedAd` | An item posted for sale by a seller | `sellerId` not empty; `title` not empty and ≤100 chars; `description` not empty; `price` valid `Money`; `category` valid `Category`; each photo URL not empty | `domain/classified_ad.go` |

**Value Objects**

| Value Object | Description | Validation Rules | Link |
|--------------|-------------|-------------------|------|
| `Money` | Price amount (cents, int64) + `Currency` | amount ≥ 0; currency must be valid | `domain/classified_ad.go` |
| `Currency` | Enum: `EUR` | must be one of the valid currencies (`EUR` only) | `domain/classified_ad.go` |
| `Category` | Enum: `Vehicles`, `RealEstate`, `Electronics`, `Furniture`, `Fashion`, `Other` | must be one of the enum values | `domain/classified_ad.go` |
| `Photo` | Photo URL attached to an ad | URL not empty | `domain/classified_ad.go` |

**Errors**: `ErrEmptySellerId`, `ErrEmptyTitle`, `ErrTitleTooLong`, `ErrEmptyDescription`, `ErrNegativeAmount`, `ErrInvalidCurrency`, `ErrInvalidCategory`, `ErrEmptyPhotoURL`, `ErrClassifiedAdNotFound` (`domain/classified_ad.go`).

### Domain Events

| Event | EventType() | When Emitted | Consumers | Link |
|-------|-------------|---------------|-----------|------|
| `ClassifiedAdPostedEvent` | `"ClassifiedAdPosted"` | After a `ClassifiedAd` is successfully saved via `PostClassifiedAdCommand` | None yet (no consumers implemented in any bounded context) | `domain/event.go` |

### Repository Ports

| Interface | Methods | Implementations | Link |
|-----------|---------|-------------------|------|
| `ClassifiedAdRepository` | `Save(classifiedAd *ClassifiedAd) error`, `GetById(id string) (*ClassifiedAd, error)` | inmemory | `domain/repository.go` |

---

## Application

### Commands

| Command | Input | Output | Emits | Link |
|---------|-------|--------|-------|------|
| `PostClassifiedAdCommand` (built by `BuildPostClassifiedAdCommand`) | `PostClassifiedAdCommandArgs{SellerId, Title, Description, PriceAmount int64, PriceCurrency, Category string, PhotoURLs []string}` | `(string, error)` — the new ad's id | `ClassifiedAdPostedEvent` | `application/command/post_classified_ad.go` |

### Queries

None implemented yet.

---

## Adapters

### Driven (Secondary)

> Implement domain ports to interact with external systems.

| Adapter | Implements | Description | Link |
|---------|------------|--------------|------|
| `InMemoryClassifiedAdRepository` | `ClassifiedAdRepository` | Thread-safe (`sync.RWMutex`) in-memory map keyed by ad id | `adapter/driven/inmemory/repository.go` |

### Driving (Primary)

> Entry points that invoke application use cases.

| Adapter | Type | Invokes | Endpoint/Trigger | Link |
|---------|------|---------|-------------------|------|
| `Handler.PostClassifiedAd` | HTTP | `PostClassifiedAdCommand` | `POST /classified-ads` | `adapter/driving/http/handler.go` |

**Request/Response DTOs** (`adapter/driving/http/dto.go`): `PostClassifiedAdRequest`, `PostClassifiedAdResponse`, `ClassifiedAdViewResponse` (unused so far — no query/handler returns it yet), `ErrorResponse`.

Domain validation errors (`ErrEmptySellerId`, `ErrEmptyTitle`, `ErrTitleTooLong`, `ErrEmptyDescription`, `ErrNegativeAmount`, `ErrInvalidCurrency`, `ErrInvalidCategory`, `ErrEmptyPhotoURL`) are mapped to HTTP 400; anything else maps to HTTP 500.

Wired in `cmd/api/main.go`: `eventbus.NewAsyncInMemoryEventBus()` → `InMemoryClassifiedAdRepository` → `BuildPostClassifiedAdCommand` → `NewHandler` → registered on `/classified-ads` via `http.ServeMux`, served on `:8080`.

---

## Dependencies

### Consumes Events From

None.

### Emits Events To

| Event | Consumed By |
|-------|-------------|
| `ClassifiedAdPostedEvent` | None yet — no consumers implemented in any bounded context |

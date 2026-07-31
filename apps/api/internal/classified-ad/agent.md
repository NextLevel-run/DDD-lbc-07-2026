# Classified Ad

**Always** add 🏷️ to STARTER_CHARACTER

**Type:** Core Domain
**Description:** Manages the lifecycle of second-hand item listings — submission, search, viewing, buyer offers, deletion and expiration.

**Status:** Implemented

---

## Ubiquitous Language

### **ClassifiedAd**
The aggregate root representing a second-hand item listed for sale by a Seller. It carries a title, description, price, image URLs, category, location, submission date and a lifecycle Status. A ClassifiedAd is created already published (`NewClassifiedAd` immediately sets `StatusPublished`) — there is no draft state. `isOnline` is a stored field mirroring `status == published`, recomputed by the internal `setStatus` helper on every transition (`NewClassifiedAd`, `Delete`, `Expire`) so it can never drift from `status`.

### **Seller**
An internal entity embedded in ClassifiedAd, identifying the person selling the item: an `Email`, a `pseudo` (display name, 1-30 chars) and a `Password` (hashed). The seller's email and password are required to authenticate deletion requests.

### **Buyer**
Not a persisted entity — a buyer is an anonymous visitor identified only by the email/pseudo they submit when making an Offer. No Buyer aggregate exists in this bounded context.

### **Offer**
A buyer's proposal to purchase a ClassifiedAd at a given amount, with an accompanying message. Offers are not persisted as entities: `MakeOfferCommand` validates the offer (message non-empty and ≤1000 chars, amount ≥0) and emits a `BuyerOfferMadeEvent` directly — no `Offer` aggregate or repository exists; the offer only exists as an event.

### **Status**
The lifecycle state of a ClassifiedAd: `published` (visible, searchable, can receive offers), `deleted` (removed by the seller), `expired` (automatically retired after `AdLifetime`). Transitions are one-way: `published → deleted` or `published → expired`; `deleted` and `expired` are terminal.

### **AdLifetime**
A domain constant (90 days) defining how long a published ad stays online before becoming expirable. `IsExpirable(now)` is true once `now >= publishedAt + AdLifetime` and the ad is still `published`.

### **DeleteReason**
Why a seller removed their ad: `sold`, `no_more_to_sell`, or `edit` (the seller intends to resubmit a corrected ad — there is no in-place edit operation, only delete + resubmit).

### **Category**
A closed enumeration of listing categories: `immo`, `auto`, `consumer_goods`, `holidays`.

### **Location**
A value object pairing a 5-digit `zipCode` with a `cityName`, used both to describe a ClassifiedAd's location and as a Search filter.

### **Price**
A value object wrapping a monetary amount in cents (`amountInCents`), always non-negative.

### **SearchCriteria**
The filter/sort/pagination contract used by the `Search` repository port: optional Category, ZipCode, CityName, price range, free-text Keywords (matched against title/description), an `OnlineOnly` flag, a `SortBy` mode (`date_desc` default, `date_asc`, `price_asc`, `price_desc`), and `Limit`/`Offset` pagination. `OnlineOnly` is set to `true` by `SearchClassifiedAdsQuery` itself — it is not a caller-supplied filter (not in `SearchClassifiedAdsQueryArgs`, not an HTTP query param) — so search only ever returns `published` ads. Any `ClassifiedAdRepository` implementation must honor `OnlineOnly` rather than hardcode its own visibility rule.

---

## User Journeys

### Submit a Classified Ad
1. Seller submits title, description, price, images, category, location and their email/pseudo/password via `POST /classified-ads`.
2. `SubmitClassifiedAdCommand` validates each value object (email, password, seller, category, location) and the ad itself, hashing the password via `PasswordHasher`.
3. The ad is persisted with `StatusPublished` and `publishedAt = submissionDate`.
4. A `ClassifiedAdPublishedEvent` is emitted after successful save.
5. `AdPublishedEmailConsumer` sends a confirmation email to the seller.

### Search Classified Ads
1. A visitor queries `GET /classified-ads` with optional filters (category, zip, city, keywords, price range), sort and pagination.
2. `SearchClassifiedAdsQuery` builds a `SearchCriteria` with `OnlineOnly: true` and delegates to the repository, which excludes non-`published` ads accordingly.
3. A paginated list of `ClassifiedAdListItemView` (summary fields + first image) is returned.

### View a Classified Ad
1. A visitor requests `GET /classified-ads/{id}`.
2. `GetClassifiedAdQuery` fetches the ad by id and returns `ErrClassifiedAdNotFound` if it doesn't exist or isn't `published` (deleted/expired ads are not viewable).
3. The full `ClassifiedAdView` (all detail fields) is returned.

### Make an Offer
1. A buyer submits an amount, message, email and pseudo via `POST /classified-ads/{id}/offers`.
2. `MakeOfferCommand` loads the ad, checks `CanReceiveOffer()` (must be `published`), validates the buyer email/message/amount.
3. No state changes on the ad; a `BuyerOfferMadeEvent` is emitted directly.
4. `OfferEmailConsumer` emails the seller with the buyer's offer details.

### Delete a Classified Ad
1. The seller sends email, password and a `DeleteReason` via `DELETE /classified-ads/{id}`.
2. `DeleteClassifiedAdCommand` loads the ad and calls `ad.Delete(...)`, which verifies the email matches and the password matches the stored hash, returning `ErrInvalidCredentials` otherwise.
3. If already `deleted`, the call is a no-op (idempotent, no error, no event).
4. Otherwise the ad transitions to `StatusDeleted`, is saved, and a `ClassifiedAdDeletedEvent` is emitted.

### Expire Outdated Ads
1. A background ticker (every hour, in `cmd/api/main.go`) invokes `ExpireOutdatedAdsCommand`.
2. The command fetches all expirable ads (`published` and past `AdLifetime`) via `FindExpirable(now)`.
3. Each eligible ad is transitioned to `StatusExpired`, saved, and emits a `ClassifiedAdExpiredEvent`.
4. Returns the count of expired ads.

---

## Architecture

### Folder Structure

```
internal/classified-ad/
├── domain/
│   ├── classified_ad.go     # Aggregate root
│   ├── seller.go            # Internal entity
│   ├── location.go          # Internal entity
│   ├── email.go             # Value object
│   ├── password.go          # Value object
│   ├── price.go             # Value object
│   ├── submission_date.go   # Value object
│   ├── category.go          # Enum value object
│   ├── status.go            # Enum
│   ├── delete_reason.go     # Enum value object
│   ├── offer.go             # Offer validation functions
│   ├── repository.go        # Repository + PasswordHasher + Clock ports
│   ├── event.go             # Domain events
│   └── errors.go            # Domain errors
├── application/
│   ├── command/
│   │   ├── submit_classified_ad.go
│   │   ├── make_offer.go
│   │   ├── delete_classified_ad.go
│   │   └── expire_outdated_ads.go
│   └── query/
│       ├── get_classified_ad.go
│       └── search_classified_ads.go
└── adapter/
    ├── driven/
    │   ├── inmemory/         # ClassifiedAdRepository implementation
    │   ├── bcrypt/           # PasswordHasher implementation
    │   └── clock/            # Clock implementations (system, fixed)
    └── driving/
        ├── http/             # HTTP handler, DTOs
        └── consumer/         # Email event consumers
```

---

## Domain

### Aggregates

| Aggregate | Description | Validation Rules | Link |
|-----------|-------------|-------------------|------|
| `ClassifiedAd` | Second-hand item listing; owns its lifecycle (published/deleted/expired) | Title non-empty & ≤100 chars; description non-empty & ≤4000 chars; price ≥0; ≤10 image URLs, none empty; delete requires matching email/password | `domain/classified_ad.go` |

### Domain Events

| Event | EventType() | When Emitted | Consumers | Link |
|-------|-------------|---------------|-----------|------|
| `ClassifiedAdPublishedEvent` | `"ClassifiedAdPublished"` | After a new ad is saved by `SubmitClassifiedAdCommand` | `AdPublishedEmailConsumer` | `domain/event.go` |
| `BuyerOfferMadeEvent` | `"BuyerOfferMade"` | When `MakeOfferCommand` accepts a valid offer on a published ad | `OfferEmailConsumer` | `domain/event.go` |
| `ClassifiedAdDeletedEvent` | `"ClassifiedAdDeleted"` | After a seller successfully deletes their ad | none (no consumer registered) | `domain/event.go` |
| `ClassifiedAdExpiredEvent` | `"ClassifiedAdExpired"` | After an ad is auto-expired by `ExpireOutdatedAdsCommand` | none (no consumer registered) | `domain/event.go` |

### Repository Ports

| Interface | Methods | Implementations | Link |
|-----------|---------|-------------------|------|
| `ClassifiedAdRepository` | `Save`, `FindByID`, `FindExpirable`, `Search` | inmemory | `domain/repository.go` |
| `PasswordHasher` | `Hash`, `Compare` | bcrypt | `domain/repository.go` |
| `Clock` | `Now` | system, fixed | `domain/repository.go` |

---

## Application

### Commands

| Command | Input | Output | Emits | Link |
|---------|-------|--------|-------|------|
| `SubmitClassifiedAdCommand` | `SubmitClassifiedAdCommandArgs` (title, description, price, seller info, images, category, location) | ad ID (`string`) | `ClassifiedAdPublishedEvent` | `application/command/submit_classified_ad.go` |
| `MakeOfferCommand` | `MakeOfferCommandArgs` (ad ID, buyer email/pseudo, amount, message) | `error` | `BuyerOfferMadeEvent` | `application/command/make_offer.go` |
| `DeleteClassifiedAdCommand` | `DeleteClassifiedAdCommandArgs` (ad ID, email, password, reason) | `error` | `ClassifiedAdDeletedEvent` | `application/command/delete_classified_ad.go` |
| `ExpireOutdatedAdsCommand` | none | count of expired ads (`int`), `error` | `ClassifiedAdExpiredEvent` (per ad) | `application/command/expire_outdated_ads.go` |

### Queries

| Query | Input | Output | Link |
|-------|-------|--------|------|
| `GetClassifiedAdQuery` | ad ID (`string`) | `ClassifiedAdView` | `application/query/get_classified_ad.go` |
| `SearchClassifiedAdsQuery` | `SearchClassifiedAdsQueryArgs` (filters, sort, pagination) | `[]ClassifiedAdListItemView` | `application/query/search_classified_ads.go` |

---

## Adapters

### Driven (Secondary)

> Implement domain ports to interact with external systems.

| Adapter | Implements | Description | Link |
|---------|------------|--------------|------|
| `InMemoryClassifiedAdRepository` | `ClassifiedAdRepository` | Thread-safe in-memory store with search filtering, sorting and pagination | `adapter/driven/inmemory/classified_ad_repository.go` |
| `BcryptPasswordHasher` | `PasswordHasher` | Hashes/compares passwords using bcrypt default cost | `adapter/driven/bcrypt/password_hasher.go` |
| `SystemClock` | `Clock` | Returns real system time | `adapter/driven/clock/system_clock.go` |
| `FixedClock` | `Clock` | Returns a fixed time (used in tests) | `adapter/driven/clock/fixed_clock.go` |

### Driving (Primary)

> Entry points that invoke application use cases.

| Adapter | Type | Invokes | Endpoint/Trigger | Link |
|---------|------|---------|--------------------|------|
| `Handler.SubmitClassifiedAd` | HTTP | `SubmitClassifiedAdCommand` | `POST /classified-ads` | `adapter/driving/http/handler.go` |
| `Handler.SearchClassifiedAds` | HTTP | `SearchClassifiedAdsQuery` | `GET /classified-ads` | `adapter/driving/http/handler.go` |
| `Handler.GetClassifiedAd` | HTTP | `GetClassifiedAdQuery` | `GET /classified-ads/{id}` | `adapter/driving/http/handler.go` |
| `Handler.MakeOffer` | HTTP | `MakeOfferCommand` | `POST /classified-ads/{id}/offers` | `adapter/driving/http/handler.go` |
| `Handler.DeleteClassifiedAd` | HTTP | `DeleteClassifiedAdCommand` | `DELETE /classified-ads/{id}` | `adapter/driving/http/handler.go` |
| `NewAdPublishedEmailConsumer` | Consumer | sends confirmation email | on `ClassifiedAdPublished` event | `adapter/driving/consumer/ad_published_email_consumer.go` |
| `NewOfferEmailConsumer` | Consumer | sends offer notification email | on `BuyerOfferMade` event | `adapter/driving/consumer/offer_email_consumer.go` |

> `ExpireOutdatedAdsCommand` is also driven, but not via an adapter in this module — it is invoked directly by an hourly `time.Ticker` goroutine in `cmd/api/main.go`.

---

## Dependencies

### Consumes Events From
| Event | Source Domain | Handler |
|-------|-----------------|---------|
| — | none | This bounded context consumes no events from other domains. |

### Emits Events To
| Event | Consumed By |
|-------|--------------|
| `ClassifiedAdPublished` | `AdPublishedEmailConsumer` (this domain) |
| `BuyerOfferMade` | `OfferEmailConsumer` (this domain) |
| `ClassifiedAdDeleted` | no consumer registered |
| `ClassifiedAdExpired` | no consumer registered |

---

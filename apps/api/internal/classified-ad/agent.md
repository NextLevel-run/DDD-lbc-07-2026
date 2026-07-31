# Classified Ad

**Type:** Core Domain
**Description:** Manages the lifecycle of second-hand item listings — submission, moderation transitions, publication, search, viewing, buyer offers, edition, deletion and expiration.

**Status:** Implemented

---

## Ubiquitous Language

### **ClassifiedAd**
The aggregate root representing a second-hand item listed for sale by a Seller. It carries a title, description, price, image URLs, category, location, submission date and a lifecycle Status. A ClassifiedAd is created in **`StatusSubmitted`**, awaiting moderation — `publishedAt` stays unset (zero time) until the ad transitions from `approved` to `published`. `isOnline` is a stored field mirroring `status == published`, recomputed by the internal `setStatus` helper on every transition so it can never drift from `status`.

### **Seller**
An internal entity embedded in ClassifiedAd, identifying the person selling the item: an `Email`, a `pseudo` (display name, 1-30 chars) and a `Password` (hashed). The seller's email and password are required to authenticate deletion and edition requests.

### **Buyer**
Not a persisted entity — a buyer is an anonymous visitor identified only by the email/pseudo they submit when making an Offer. No Buyer aggregate exists in this bounded context.

### **Offer**
A buyer's proposal to purchase a ClassifiedAd at a given amount, with an accompanying message. Offers are not persisted as entities: `MakeOfferCommand` validates the offer (message non-empty and ≤1000 chars, amount ≥0) and emits a `BuyerOfferMadeEvent` directly — no `Offer` aggregate or repository exists; the offer only exists as an event.

### **Status**
The lifecycle state of a ClassifiedAd: `submitted` (awaiting moderation), `approved` (accepted by moderation, not yet online), `challenged` (moderation asked the seller for corrections), `rejected` (refused by moderation, immediately auto-deleted), `published` (visible, searchable, can receive offers), `deleted` (removed), `expired` (automatically retired after `AdLifetime`). Full lifecycle:

```
submitted → approved → published → deleted | expired
submitted → rejected → deleted (automatic, reason "rejected")
submitted → challenged → (edit) → submitted (re-submission)
```

The moderation transitions (`Approve()`, `Publish(now)`, `Reject()`, `Challenge()`, `DeleteRejected(now)`) are driven by system commands with no seller credentials; each one guards its source status (`ErrCannotApprove`, `ErrCannotPublish`, `ErrCannotReject`, `ErrCannotChallenge`, `ErrCannotDeleteRejected`). `deleted` and `expired` are terminal. Only `published` ads are online.

### **AdLifetime**
A domain constant (90 days) defining how long a published ad stays online before becoming expirable. `IsExpirable(now)` is true once `now >= publishedAt + AdLifetime` and the ad is still `published`.

### **DeleteReason**
Why an ad was deleted: `sold`, `no_more_to_sell`, `edit` (the legacy delete+resubmit workaround, which coexists with the real in-place `Edit` on challenged ads) or `rejected` (set automatically by the system when moderation rejects the ad — never chosen by a seller).

### **Edit**
The in-place correction of a **challenged** ad by its seller: `Edit(...)` re-validates all editable content (title, description, price, images, category, location) with the same rules as `NewClassifiedAd` and re-submits the ad (`challenged → submitted`), triggering a fresh moderation round. Guarded by `ErrCannotEdit` when the ad is not challenged.

### **Category**
A closed enumeration of listing categories: `immo`, `auto`, `consumer_goods`, `holidays`.

### **Location**
A value object pairing a 5-digit `zipCode` with a `cityName`, used both to describe a ClassifiedAd's location and as a Search filter.

### **Price**
A value object wrapping a monetary amount in cents (`amountInCents`), always non-negative.

### **SearchCriteria**
The filter/sort/pagination contract used by the `Search` repository port: optional Category, ZipCode, CityName, price range, free-text Keywords (matched against title/description), an `OnlineOnly` flag, a `SortBy` mode (`date_desc` default, `date_asc`, `price_asc`, `price_desc`), and `Limit`/`Offset` pagination. `OnlineOnly` is set to `true` by `SearchClassifiedAdsQuery` itself — it is not a caller-supplied filter (not in `SearchClassifiedAdsQueryArgs`, not an HTTP query param) — so search only ever returns `published` ads: `submitted`, `approved`, `challenged` and `rejected` ads never appear in public listings. Any `ClassifiedAdRepository` implementation must honor `OnlineOnly` rather than hardcode its own visibility rule.

### **Public events (integration contracts)**
This context communicates with Moderation only through the public bus and the DTOs in `internal/shared`: publisher adapters relay the internal `Submitted`, `Edited`, `Published`, `Deleted` and `Expired` events to the public bus, and consumers react to the public Moderation decisions `ClassifiedAdApproved`, `ClassifiedAdRejected` and `ClassifiedAdChallenged`. Internal domain events never cross the context boundary.

---

## User Journeys

### Submit a Classified Ad
1. Seller submits title, description, price, images, category, location and their email/pseudo/password via `POST /classified-ads`.
2. `SubmitClassifiedAdCommand` validates each value object (email, password, seller, category, location) and the ad itself, hashing the password via `PasswordHasher`.
3. The ad is persisted with `StatusSubmitted`, `publishedAt` unset.
4. A `ClassifiedAdSubmittedEvent` (full ad payload) is emitted after successful save; the publisher relays it as the public `ClassifiedAdSubmitted`, which creates a ModerationTask in the Moderation context.

### Moderation Approval (chained publication)
1. Moderation emits the public `ClassifiedAdApproved`.
2. `NewClassifiedAdApprovedConsumer` (public bus) calls `ApproveClassifiedAdCommand`: `submitted → approved`, internal `ClassifiedAdApprovedEvent` emitted.
3. `NewClassifiedAdApprovedInternalConsumer` (internal bus) reacts immediately and calls `PublishClassifiedAdCommand`: `approved → published`, `publishedAt` set via `Clock`, `ClassifiedAdPublishedEvent` emitted.
4. `AdPublishedEmailConsumer` sends a confirmation email to the seller; the publisher relays the public `ClassifiedAdPublished`.

### Moderation Rejection (automatic deletion)
1. Moderation emits the public `ClassifiedAdRejected` (with a reason).
2. `NewClassifiedAdRejectedConsumer` calls `RejectClassifiedAdCommand` — a system command without seller credentials that applies `Reject()` (`submitted → rejected`) then `DeleteRejected(now)` (`rejected → deleted`, reason `rejected`), emitting the internal `ClassifiedAdRejectedEvent` then `ClassifiedAdDeletedEvent`.
3. The ad never appears in public queries.

### Moderation Challenge (seller correction loop)
1. Moderation emits the public `ClassifiedAdChallenged` (with a reason).
2. `NewClassifiedAdChallengedConsumer` calls `ChallengeClassifiedAdCommand` (`submitted → challenged`, internal `ClassifiedAdChallengedEvent`), then emails the seller (via `pkg/mailer`, looking the seller up through the repository since the public event does not carry it).
3. The seller corrects the ad via `PUT /classified-ads/{id}` (see Edit journey), which re-submits it for a fresh moderation round.

### Edit a Challenged Classified Ad
1. The seller sends email, password and the full new content (title, description, price, images, category, location) via `PUT /classified-ads/{id}`.
2. `EditClassifiedAdCommand` authenticates the seller (same email + password mechanism as Delete, `ErrInvalidCredentials` otherwise), then calls `ad.Edit(...)`, which validates the content with the same rules as `NewClassifiedAd` and transitions `challenged → submitted` (`ErrCannotEdit` if not challenged).
3. A `ClassifiedAdEditedEvent` (full ad payload) is emitted; the public `ClassifiedAdEdited` creates a **new** ModerationTask.

### Search Classified Ads
1. A visitor queries `GET /classified-ads` with optional filters (category, zip, city, keywords, price range), sort and pagination.
2. `SearchClassifiedAdsQuery` builds a `SearchCriteria` with `OnlineOnly: true` and delegates to the repository, which excludes non-`published` ads accordingly.
3. A paginated list of `ClassifiedAdListItemView` (summary fields + first image) is returned.

### View a Classified Ad
1. A visitor requests `GET /classified-ads/{id}`.
2. `GetClassifiedAdQuery` fetches the ad by id and returns `ErrClassifiedAdNotFound` if it doesn't exist or isn't online (non-`published` ads — submitted, approved, challenged, rejected, deleted, expired — are not viewable).
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
│   ├── classified_ad.go     # Aggregate root (+ moderation transitions, Edit)
│   ├── seller.go            # Internal entity
│   ├── location.go          # Internal entity
│   ├── email.go             # Value object
│   ├── password.go          # Value object
│   ├── price.go             # Value object
│   ├── submission_date.go   # Value object
│   ├── category.go          # Enum value object
│   ├── status.go            # Enum (7 statuses)
│   ├── delete_reason.go     # Enum value object (incl. rejected)
│   ├── offer.go             # Offer validation functions
│   ├── repository.go        # Repository + PasswordHasher + Clock ports
│   ├── event.go             # Domain events
│   └── errors.go            # Domain errors
├── application/
│   ├── command/
│   │   ├── submit_classified_ad.go
│   │   ├── edit_classified_ad.go       # seller, challenged → submitted
│   │   ├── approve_classified_ad.go    # system, submitted → approved
│   │   ├── publish_classified_ad.go    # system, approved → published
│   │   ├── reject_classified_ad.go     # system, submitted → rejected → deleted
│   │   ├── challenge_classified_ad.go  # system, submitted → challenged
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
        ├── consumer/         # Moderation-decision + email consumers (one per file)
        └── publisher/        # internal events → public bus (one per file)
```

---

## Domain

### Aggregates

| Aggregate | Description | Validation Rules | Link |
|-----------|-------------|-------------------|------|
| `ClassifiedAd` | Second-hand item listing; owns its full lifecycle (submitted/approved/challenged/rejected/published/deleted/expired) | Title non-empty & ≤100 chars; description non-empty & ≤4000 chars; price ≥0; ≤10 image URLs, none empty (shared by `NewClassifiedAd` and `Edit` via `validateContent`); delete/edit require matching email/password; every transition guards its source status | `domain/classified_ad.go` |

### Domain Events

All events below are emitted on the **internal** ClassifiedAd bus. `Submitted`, `Edited`, `Published`, `Deleted` and `Expired` are additionally relayed to the public bus by the publisher adapters (as the DTOs in `internal/shared`).

| Event | EventType() | When Emitted | Consumers | Link |
|-------|-------------|---------------|-----------|------|
| `ClassifiedAdSubmittedEvent` | `"ClassifiedAdSubmitted"` | After a new ad is saved by `SubmitClassifiedAdCommand` (full ad payload) | `NewClassifiedAdSubmittedPublisher` → public bus | `domain/event.go` |
| `ClassifiedAdEditedEvent` | `"ClassifiedAdEdited"` | After a challenged ad is edited and re-submitted by `EditClassifiedAdCommand` (full ad payload) | `NewClassifiedAdEditedPublisher` → public bus | `domain/event.go` |
| `ClassifiedAdApprovedEvent` | `"ClassifiedAdApproved"` | After `ApproveClassifiedAdCommand` (submitted → approved) | `NewClassifiedAdApprovedInternalConsumer` (chains publication) | `domain/event.go` |
| `ClassifiedAdPublishedEvent` | `"ClassifiedAdPublished"` | After `PublishClassifiedAdCommand` puts an approved ad online | `AdPublishedEmailConsumer`, `NewClassifiedAdPublishedPublisher` → public bus | `domain/event.go` |
| `ClassifiedAdRejectedEvent` | `"ClassifiedAdRejected"` | After `RejectClassifiedAdCommand` (submitted → rejected), before the deleted event | none registered | `domain/event.go` |
| `ClassifiedAdChallengedEvent` | `"ClassifiedAdChallenged"` | After `ChallengeClassifiedAdCommand` (submitted → challenged) | none registered | `domain/event.go` |
| `BuyerOfferMadeEvent` | `"BuyerOfferMade"` | When `MakeOfferCommand` accepts a valid offer on a published ad | `OfferEmailConsumer` | `domain/event.go` |
| `ClassifiedAdDeletedEvent` | `"ClassifiedAdDeleted"` | After a seller deletion, or automatically after a rejection (reason `rejected`) | `NewClassifiedAdDeletedPublisher` → public bus | `domain/event.go` |
| `ClassifiedAdExpiredEvent` | `"ClassifiedAdExpired"` | After an ad is auto-expired by `ExpireOutdatedAdsCommand` | `NewClassifiedAdExpiredPublisher` → public bus | `domain/event.go` |

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
| `SubmitClassifiedAdCommand` | `SubmitClassifiedAdCommandArgs` (title, description, price, seller info, images, category, location) | ad ID (`string`) | `ClassifiedAdSubmittedEvent` | `application/command/submit_classified_ad.go` |
| `EditClassifiedAdCommand` | `EditClassifiedAdCommandArgs` (ad ID, email, password, full new content) | `error` | `ClassifiedAdEditedEvent` | `application/command/edit_classified_ad.go` |
| `ApproveClassifiedAdCommand` | `ApproveClassifiedAdCommandArgs` (ad ID) — system, no credentials | `error` | `ClassifiedAdApprovedEvent` | `application/command/approve_classified_ad.go` |
| `PublishClassifiedAdCommand` | `PublishClassifiedAdCommandArgs` (ad ID) — system, no credentials | `error` | `ClassifiedAdPublishedEvent` | `application/command/publish_classified_ad.go` |
| `RejectClassifiedAdCommand` | `RejectClassifiedAdCommandArgs` (ad ID) — system, no credentials | `error` | `ClassifiedAdRejectedEvent` then `ClassifiedAdDeletedEvent` | `application/command/reject_classified_ad.go` |
| `ChallengeClassifiedAdCommand` | `ChallengeClassifiedAdCommandArgs` (ad ID) — system, no credentials | `error` | `ClassifiedAdChallengedEvent` | `application/command/challenge_classified_ad.go` |
| `MakeOfferCommand` | `MakeOfferCommandArgs` (ad ID, buyer email/pseudo, amount, message) | `error` | `BuyerOfferMadeEvent` | `application/command/make_offer.go` |
| `DeleteClassifiedAdCommand` | `DeleteClassifiedAdCommandArgs` (ad ID, email, password, reason) | `error` | `ClassifiedAdDeletedEvent` | `application/command/delete_classified_ad.go` |
| `ExpireOutdatedAdsCommand` | none | count of expired ads (`int`), `error` | `ClassifiedAdExpiredEvent` (per ad) | `application/command/expire_outdated_ads.go` |

> The four system commands (`Approve`, `Publish`, `Reject`, `Challenge`) are driven only by event consumers — they have no HTTP endpoint and take no seller credentials. `DeleteClassifiedAdCommand` remains the seller-facing deletion; the rejection chain uses the dedicated `RejectClassifiedAdCommand` instead.

### Queries

| Query | Input | Output | Link |
|-------|-------|--------|------|
| `GetClassifiedAdQuery` | ad ID (`string`) | `ClassifiedAdView` (online ads only) | `application/query/get_classified_ad.go` |
| `SearchClassifiedAdsQuery` | `SearchClassifiedAdsQueryArgs` (filters, sort, pagination) | `[]ClassifiedAdListItemView` (published ads only) | `application/query/search_classified_ads.go` |

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
| `Handler.EditClassifiedAd` | HTTP | `EditClassifiedAdCommand` | `PUT /classified-ads/{id}` | `adapter/driving/http/handler.go` |
| `NewClassifiedAdApprovedConsumer` | Consumer | `ApproveClassifiedAdCommand` | public `ClassifiedAdApproved` | `adapter/driving/consumer/classified_ad_approved_consumer.go` |
| `NewClassifiedAdApprovedInternalConsumer` | Consumer | `PublishClassifiedAdCommand` (chained publication) | internal `ClassifiedAdApproved` | `adapter/driving/consumer/classified_ad_approved_internal_consumer.go` |
| `NewClassifiedAdRejectedConsumer` | Consumer | `RejectClassifiedAdCommand` (auto-deletion, reason `rejected`) | public `ClassifiedAdRejected` | `adapter/driving/consumer/classified_ad_rejected_consumer.go` |
| `NewClassifiedAdChallengedConsumer` | Consumer | `ChallengeClassifiedAdCommand` + correction email to the seller | public `ClassifiedAdChallenged` | `adapter/driving/consumer/classified_ad_challenged_consumer.go` |
| `NewAdPublishedEmailConsumer` | Consumer | sends confirmation email | internal `ClassifiedAdPublished` | `adapter/driving/consumer/ad_published_email_consumer.go` |
| `NewOfferEmailConsumer` | Consumer | sends offer notification email | internal `BuyerOfferMade` | `adapter/driving/consumer/offer_email_consumer.go` |
| `RegisterPublishers` | Publisher | relays internal events to the public bus | internal `Submitted`/`Edited`/`Published`/`Deleted`/`Expired` events | `adapter/driving/publisher/register.go` |

> `ExpireOutdatedAdsCommand` is also driven, but not via an adapter in this module — it is invoked directly by an hourly `time.Ticker` goroutine in `cmd/api/main.go`.

---

## Dependencies

### Consumes Events From
| Event (public bus) | Source Domain | Handler |
|-------|-----------------|---------|
| `ClassifiedAdApproved` | Moderation | `NewClassifiedAdApprovedConsumer` (approve, then publish via the internal chain) |
| `ClassifiedAdRejected` | Moderation | `NewClassifiedAdRejectedConsumer` (reject + automatic deletion) |
| `ClassifiedAdChallenged` | Moderation | `NewClassifiedAdChallengedConsumer` (challenge + seller email) |

### Emits Events To
| Event (public bus) | Consumed By |
|-------|--------------|
| `ClassifiedAdSubmitted` | Moderation (creates a ModerationTask + `submitted` history entry with snapshot) |
| `ClassifiedAdEdited` | Moderation (creates a new ModerationTask + `submitted` history entry with snapshot) |
| `ClassifiedAdPublished` | Moderation (`published` history entry) |
| `ClassifiedAdDeleted` | Moderation (`deleted` history entry) |
| `ClassifiedAdExpired` | Moderation (`expired` history entry) |

> Contracts live in `internal/shared` (public event DTOs + `PublicEventBus`, an alias of `pkg/eventbus.Bus`). Internal domain events (`BuyerOfferMade`, the internal `Approved`/`Rejected`/`Challenged`, …) never cross the context boundary — only the public DTOs relayed by the publisher adapters do.

---

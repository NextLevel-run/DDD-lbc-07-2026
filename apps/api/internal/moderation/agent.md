# Moderation

**Type:** Supporting Domain
**Description:** Reviews every submitted (and re-submitted) classified ad before it goes online — task queue with exclusive claims, accept/reject/challenge decisions with predefined reasons, and a full append-only audit trail per ad.

**Status:** Implemented

---

## Ubiquitous Language

### **Moderator**
An aggregate root representing a person reviewing classified ads: an `id` (uuid) and a `fullName` kept for traceability. No authentication is handled in this scope — the moderator is considered already authenticated upstream; `moderatorId` is simply passed in the JSON body of moderation HTTP requests. Two moderators with well-known fixed IDs are seeded at startup in `cmd/api/main.go` (via `RehydrateModerator`).

### **ModerationTask**
The aggregate root representing one classified ad awaiting a moderation decision. It carries an `id`, `createdAt`, the `classifiedAdID` under review, and a nullable `moderatorID`/`claimedAt` pair (nil while unclaimed). Tasks live in a shared queue visible to all moderators. Lifecycle: **Created → Claimed → Completed (physically deleted)**. Each submission or re-submission of an ad creates a **new** task with a new ID.

### **Claim**
An exclusive lock on a task: `Claim(moderatorID, now)` fails with `ErrTaskAlreadyClaimed` if any moderator (including the same one) already holds the task. There is no claim timeout.

### **Complete**
Finishing a task with a decision (accept, reject or challenge). `Complete(moderatorID)` only verifies ownership — `ErrNotTaskOwner` unless the caller is the moderator holding the claim (an unclaimed task has no owner, so it can never be completed). The task is then **physically deleted** by the command (the audit trail lives in ClassifiedAdHistory, not in the task) and a `ModerationTaskCompletedEvent` is emitted alongside the decision event.

### **ClassifiedAdHistory**
An aggregate root recording the full moderation history of one classified ad as an **append-only event log** (`classifiedAdID` + chronological `entries`). Entries are never updated nor removed. The ad's current status as known by this context is derived from the last entry (`CurrentStatus()`), never stored. `LastSnapshot()` returns the most recent ad content snapshot in the log. It is fed exclusively by consumers of public events — never written by HTTP commands directly.

### **HistoryEntry**
One immutable record in the log: `occurredAt`, an `action`, and three optional fields depending on the action — `moderatorID` (moderation actions), `reason` (reject/challenge), `snapshot` (submitted/edited).

### **HistoryAction**
A closed enumeration of recordable events: `submitted`, `approved`, `rejected`, `challenged`, `published`, `deleted`, `expired`.

### **ClassifiedAdSnapshot**
A value object capturing the full ad content at submission/edition time (title, description, priceInCents, imageURLs, category, zipCode, cityName, sellerEmail, sellerPseudo). A pure data carrier copied from public integration events — no invariants, hence exported fields.

### **RejectReason**
A closed enumeration of predefined rejection reasons: `inappropriate_content`, `suspect_price`, `wrong_category`. Any other value fails with `ErrInvalidRejectReason`.

### **ChallengeReason**
A closed enumeration of predefined challenge reasons: `price_to_verify`, `category_to_fix`. Any other value fails with `ErrInvalidChallengeReason`.

### **Public events (integration contracts)**
This context communicates with ClassifiedAd only through the public bus and the DTOs in `internal/shared`: it **consumes** `ClassifiedAdSubmitted`, `ClassifiedAdEdited`, `ClassifiedAdPublished`, `ClassifiedAdDeleted`, `ClassifiedAdExpired` (plus its own three outbound events, re-consumed for the audit trail) and **produces** `ClassifiedAdApproved`, `ClassifiedAdRejected`, `ClassifiedAdChallenged`.

---

## User Journeys

### Task Creation (automatic)
1. A seller submits (or edits a challenged) ad in the ClassifiedAd context; the public `ClassifiedAdSubmitted` / `ClassifiedAdEdited` event reaches the public bus.
2. `NewClassifiedAdSubmittedConsumer` / `NewClassifiedAdEditedConsumer` calls `CreateModerationTaskCommand`, enqueuing a brand new task (new ID even for a re-submission).
3. The same consumer then calls `AppendHistoryEntryCommand` to record a `submitted` entry carrying the full ad snapshot.

### Browse the Moderation Queue
1. A moderator requests `GET /moderation/tasks`.
2. `ListModerationTasksQuery` returns every active task (pending and claimed), oldest first (ID tie-break for determinism): task ID, ad title (from the last history snapshot, empty if the history has not been fed yet), creation date, status (`pending`/`claimed`) and claimer full name (resolved via `ModeratorRepository`).

### Inspect a Task
1. A moderator requests `GET /moderation/tasks/{id}`.
2. `GetModerationTaskDetailQuery` returns the task, the full `ClassifiedAdHistory` (entries with action, moderatorID, reason, snapshot) and the last content snapshot. A missing history yields empty entries; an unknown task returns `ErrModerationTaskNotFound` (404).

### Claim a Task
1. A moderator posts their `moderatorId` to `POST /moderation/tasks/{id}/claim`.
2. `ClaimModerationTaskCommand` verifies the moderator exists, locks the task exclusively (`ErrTaskAlreadyClaimed` → 409 if already held) and emits a `ModerationTaskClaimedEvent`.

### Accept an Ad
1. The claiming moderator posts to `POST /moderation/tasks/{id}/accept`.
2. `AcceptClassifiedAdCommand` checks ownership (`ErrNotTaskOwner` → 403), physically deletes the task and emits `ModerationTaskCompletedEvent` then the internal `ClassifiedAdApprovedEvent`.
3. The publisher relays it as the public `ClassifiedAdApproved`; the ClassifiedAd context reacts (submitted → approved → published, chained) and this context's own consumer appends an `approved` history entry.

### Reject an Ad
1. The claiming moderator posts a valid `RejectReason` to `POST /moderation/tasks/{id}/reject`.
2. `RejectClassifiedAdCommand` validates the reason (`ErrInvalidRejectReason` → 400), checks ownership, deletes the task and emits `ModerationTaskCompletedEvent` then the internal `ClassifiedAdRejectedEvent`.
3. The public `ClassifiedAdRejected` triggers the automatic deletion of the ad in ClassifiedAd (reason `rejected`), and a `rejected` history entry here.

### Challenge an Ad
1. The claiming moderator posts a valid `ChallengeReason` to `POST /moderation/tasks/{id}/challenge`.
2. `ChallengeClassifiedAdCommand` validates the reason, checks ownership, deletes the task and emits `ModerationTaskCompletedEvent` then the internal `ClassifiedAdChallengedEvent`.
3. The public `ClassifiedAdChallenged` moves the ad to `challenged` in ClassifiedAd and emails the seller; a `challenged` history entry is appended here. When the seller edits the ad, a new task is created (back to Task Creation).

### Audit Trail
1. Every public ClassifiedAd event (`Published`, `Deleted`, `Expired`) and every public Moderation event (`Approved`, `Rejected`, `Challenged`) is consumed by its dedicated consumer, which appends the corresponding history entry via `AppendHistoryEntryCommand` — creating the history on first append.

---

## Architecture

### Folder Structure

```
internal/moderation/
├── domain/
│   ├── moderator.go              # Aggregate root
│   ├── moderation_task.go        # Aggregate root
│   ├── classified_ad_history.go  # Aggregate root (append-only log)
│   ├── history_entry.go          # HistoryEntry + HistoryAction enum
│   ├── snapshot.go               # ClassifiedAdSnapshot value object
│   ├── reject_reason.go          # Enum value object
│   ├── challenge_reason.go       # Enum value object
│   ├── repository.go             # Repository + Clock ports
│   ├── event.go                  # Domain events
│   └── errors.go                 # Domain errors
├── application/
│   ├── command/
│   │   ├── claim_moderation_task.go
│   │   ├── accept_classified_ad.go
│   │   ├── reject_classified_ad.go
│   │   ├── challenge_classified_ad.go
│   │   ├── create_moderation_task.go   # internal, consumer-driven
│   │   └── append_history_entry.go     # internal, consumer-driven
│   └── query/
│       ├── list_moderation_tasks.go
│       └── get_moderation_task_detail.go
└── adapter/
    ├── driven/
    │   └── inmemory/             # 3 repository implementations
    └── driving/
        ├── http/                 # HTTP handler, DTOs
        ├── consumer/             # 8 public-event consumers (one per file)
        └── publisher/            # internal events → public bus (one per file)
```

---

## Domain

### Aggregates

| Aggregate | Description | Validation Rules | Link |
|-----------|-------------|-------------------|------|
| `Moderator` | Person reviewing ads, kept for traceability | fullName non-empty; rehydration requires a non-nil ID | `domain/moderator.go` |
| `ModerationTask` | One ad awaiting a decision; shared queue, exclusive claim, deleted on completion | classifiedAdID non-empty; claim only if unclaimed (`ErrTaskAlreadyClaimed`); complete only by claim owner (`ErrNotTaskOwner`) | `domain/moderation_task.go` |
| `ClassifiedAdHistory` | Append-only moderation log of one ad; current status derived from the last entry | classifiedAdID non-empty; entries validated by `NewHistoryEntry` (known `HistoryAction`); never mutated after append | `domain/classified_ad_history.go` |

### Domain Events

All Moderation events are emitted on the **internal** Moderation bus. The three `ClassifiedAd*Event` decision events are relayed to the public bus by the publisher adapters; the two `ModerationTask*Event` events stay internal (no consumer registered).

| Event | EventType() | When Emitted | Consumers | Link |
|-------|-------------|---------------|-----------|------|
| `ModerationTaskClaimedEvent` | `"ModerationTaskClaimed"` | After a moderator claims a task | none registered | `domain/event.go` |
| `ModerationTaskCompletedEvent` | `"ModerationTaskCompleted"` | After a claimed task is completed and deleted (accept/reject/challenge) | none registered | `domain/event.go` |
| `ClassifiedAdApprovedEvent` | `"ClassifiedAdApproved"` | After `AcceptClassifiedAdCommand` deletes the task | `NewClassifiedAdApprovedPublisher` → public bus | `domain/event.go` |
| `ClassifiedAdRejectedEvent` | `"ClassifiedAdRejected"` | After `RejectClassifiedAdCommand` deletes the task | `NewClassifiedAdRejectedPublisher` → public bus | `domain/event.go` |
| `ClassifiedAdChallengedEvent` | `"ClassifiedAdChallenged"` | After `ChallengeClassifiedAdCommand` deletes the task | `NewClassifiedAdChallengedPublisher` → public bus | `domain/event.go` |

### Repository Ports

| Interface | Methods | Implementations | Link |
|-----------|---------|-------------------|------|
| `ModerationTaskRepository` | `Save`, `FindByID`, `FindAll`, `Delete` | inmemory | `domain/repository.go` |
| `ModeratorRepository` | `Save`, `FindByID` | inmemory | `domain/repository.go` |
| `ClassifiedAdHistoryRepository` | `Save`, `FindByClassifiedAdID` | inmemory | `domain/repository.go` |
| `Clock` | `Now` | reuses classified-ad's system/fixed clocks | `domain/repository.go` |

---

## Application

### Commands

| Command | Input | Output | Emits | Link |
|---------|-------|--------|-------|------|
| `ClaimModerationTaskCommand` | `ClaimModerationTaskCommandArgs` (taskID, moderatorID) | `error` | `ModerationTaskClaimedEvent` | `application/command/claim_moderation_task.go` |
| `AcceptClassifiedAdCommand` | `AcceptClassifiedAdCommandArgs` (taskID, moderatorID) | `error` | `ModerationTaskCompletedEvent`, `ClassifiedAdApprovedEvent` | `application/command/accept_classified_ad.go` |
| `RejectClassifiedAdCommand` | `RejectClassifiedAdCommandArgs` (taskID, moderatorID, reason) | `error` | `ModerationTaskCompletedEvent`, `ClassifiedAdRejectedEvent` | `application/command/reject_classified_ad.go` |
| `ChallengeClassifiedAdCommand` | `ChallengeClassifiedAdCommandArgs` (taskID, moderatorID, reason) | `error` | `ModerationTaskCompletedEvent`, `ClassifiedAdChallengedEvent` | `application/command/challenge_classified_ad.go` |
| `CreateModerationTaskCommand` | `CreateModerationTaskCommandArgs` (classifiedAdID) | task ID (`string`) | none | `application/command/create_moderation_task.go` |
| `AppendHistoryEntryCommand` | `AppendHistoryEntryCommandArgs` (classifiedAdID, occurredAt, action, optional moderatorID/reason/snapshot) | `error` | none | `application/command/append_history_entry.go` |

> `CreateModerationTaskCommand` and `AppendHistoryEntryCommand` are internal use cases driven only by event consumers — they have no HTTP endpoint.

### Queries

| Query | Input | Output | Link |
|-------|-------|--------|------|
| `ListModerationTasksQuery` | none | `[]ModerationTaskListItem` (id, ad title, createdAt, status `pending`/`claimed`, claimedBy) | `application/query/list_moderation_tasks.go` |
| `GetModerationTaskDetailQuery` | task ID (`string`) | `ModerationTaskDetailView` (task fields + full history + last snapshot) | `application/query/get_moderation_task_detail.go` |

---

## Adapters

### Driven (Secondary)

> Implement domain ports to interact with external systems.

| Adapter | Implements | Description | Link |
|---------|------------|--------------|------|
| `InMemoryModerationTaskRepository` | `ModerationTaskRepository` | Thread-safe in-memory task store with physical delete | `adapter/driven/inmemory/moderation_task_repository.go` |
| `InMemoryModeratorRepository` | `ModeratorRepository` | Thread-safe in-memory moderator store | `adapter/driven/inmemory/moderator_repository.go` |
| `InMemoryClassifiedAdHistoryRepository` | `ClassifiedAdHistoryRepository` | Thread-safe in-memory history store keyed by ad ID | `adapter/driven/inmemory/classified_ad_history_repository.go` |

### Driving (Primary)

> Entry points that invoke application use cases.

| Adapter | Type | Invokes | Endpoint/Trigger | Link |
|---------|------|---------|--------------------|------|
| `Handler.ListModerationTasks` | HTTP | `ListModerationTasksQuery` | `GET /moderation/tasks` | `adapter/driving/http/handler.go` |
| `Handler.GetModerationTaskDetail` | HTTP | `GetModerationTaskDetailQuery` | `GET /moderation/tasks/{id}` | `adapter/driving/http/handler.go` |
| `Handler.ClaimModerationTask` | HTTP | `ClaimModerationTaskCommand` | `POST /moderation/tasks/{id}/claim` | `adapter/driving/http/handler.go` |
| `Handler.AcceptClassifiedAd` | HTTP | `AcceptClassifiedAdCommand` | `POST /moderation/tasks/{id}/accept` | `adapter/driving/http/handler.go` |
| `Handler.RejectClassifiedAd` | HTTP | `RejectClassifiedAdCommand` | `POST /moderation/tasks/{id}/reject` | `adapter/driving/http/handler.go` |
| `Handler.ChallengeClassifiedAd` | HTTP | `ChallengeClassifiedAdCommand` | `POST /moderation/tasks/{id}/challenge` | `adapter/driving/http/handler.go` |
| `NewClassifiedAdSubmittedConsumer` | Consumer | `CreateModerationTaskCommand` + `AppendHistoryEntryCommand` | public `ClassifiedAdSubmitted` | `adapter/driving/consumer/classified_ad_submitted_consumer.go` |
| `NewClassifiedAdEditedConsumer` | Consumer | `CreateModerationTaskCommand` + `AppendHistoryEntryCommand` | public `ClassifiedAdEdited` | `adapter/driving/consumer/classified_ad_edited_consumer.go` |
| `NewClassifiedAdPublishedConsumer` | Consumer | `AppendHistoryEntryCommand` (`published`) | public `ClassifiedAdPublished` | `adapter/driving/consumer/classified_ad_published_consumer.go` |
| `NewClassifiedAdDeletedConsumer` | Consumer | `AppendHistoryEntryCommand` (`deleted`) | public `ClassifiedAdDeleted` | `adapter/driving/consumer/classified_ad_deleted_consumer.go` |
| `NewClassifiedAdExpiredConsumer` | Consumer | `AppendHistoryEntryCommand` (`expired`) | public `ClassifiedAdExpired` | `adapter/driving/consumer/classified_ad_expired_consumer.go` |
| `NewClassifiedAdApprovedConsumer` | Consumer | `AppendHistoryEntryCommand` (`approved`, moderatorID) | public `ClassifiedAdApproved` (own event, re-consumed) | `adapter/driving/consumer/classified_ad_approved_consumer.go` |
| `NewClassifiedAdRejectedConsumer` | Consumer | `AppendHistoryEntryCommand` (`rejected`, moderatorID, reason) | public `ClassifiedAdRejected` (own event, re-consumed) | `adapter/driving/consumer/classified_ad_rejected_consumer.go` |
| `NewClassifiedAdChallengedConsumer` | Consumer | `AppendHistoryEntryCommand` (`challenged`, moderatorID, reason) | public `ClassifiedAdChallenged` (own event, re-consumed) | `adapter/driving/consumer/classified_ad_challenged_consumer.go` |
| `RegisterPublishers` | Publisher | relays internal decision events to the public bus | internal `ClassifiedAdApproved`/`Rejected`/`Challenged` events | `adapter/driving/publisher/register_publishers.go` |

### HTTP Error Mapping

| Domain error | HTTP status |
|--------------|-------------|
| `ErrTaskAlreadyClaimed` | 409 Conflict |
| `ErrNotTaskOwner` | 403 Forbidden |
| `ErrModerationTaskNotFound`, `ErrModeratorNotFound` | 404 Not Found |
| `ErrInvalidRejectReason`, `ErrInvalidChallengeReason` | 400 Bad Request |
| anything else | 500 Internal Server Error |

---

## Dependencies

### Consumes Events From
| Event (public bus) | Source Domain | Handler |
|-------|-----------------|---------|
| `ClassifiedAdSubmitted` | ClassifiedAd | `NewClassifiedAdSubmittedConsumer` |
| `ClassifiedAdEdited` | ClassifiedAd | `NewClassifiedAdEditedConsumer` |
| `ClassifiedAdPublished` | ClassifiedAd | `NewClassifiedAdPublishedConsumer` |
| `ClassifiedAdDeleted` | ClassifiedAd | `NewClassifiedAdDeletedConsumer` |
| `ClassifiedAdExpired` | ClassifiedAd | `NewClassifiedAdExpiredConsumer` |
| `ClassifiedAdApproved` | Moderation (own event, re-consumed for the audit trail) | `NewClassifiedAdApprovedConsumer` |
| `ClassifiedAdRejected` | Moderation (own event, re-consumed for the audit trail) | `NewClassifiedAdRejectedConsumer` |
| `ClassifiedAdChallenged` | Moderation (own event, re-consumed for the audit trail) | `NewClassifiedAdChallengedConsumer` |

### Emits Events To
| Event (public bus) | Consumed By |
|-------|--------------|
| `ClassifiedAdApproved` | ClassifiedAd (`NewClassifiedAdApprovedConsumer` → approve then publish, chained), Moderation (history) |
| `ClassifiedAdRejected` | ClassifiedAd (`NewClassifiedAdRejectedConsumer` → automatic deletion, reason `rejected`), Moderation (history) |
| `ClassifiedAdChallenged` | ClassifiedAd (`NewClassifiedAdChallengedConsumer` → challenge + seller email), Moderation (history) |

> Contracts live in `internal/shared` (public event DTOs + `PublicEventBus`, an alias of `pkg/eventbus.Bus`). Internal domain events never cross the context boundary — only the public DTOs relayed by the publisher adapters do.

---

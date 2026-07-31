# Spec — Bounded Context Moderation

## Objectif

Implémenter le bounded context **Moderation** et l'intégrer avec **ClassifiedAd** via un système d'événements publics partagés.

---

## 1. Modifications sur ClassifiedAd (existant)

### 1.1 Nouveaux statuts

Ajouter les statuts suivants à `Status` :

| Statut | Description |
|--------|-------------|
| `submitted` | Annonce créée, en attente de modération |
| `approved` | Approuvée par la modération, pas encore publiée |
| `challenged` | Le modérateur demande des corrections au seller |
| `rejected` | Rejetée par la modération (terminal) |

Le cycle de vie complet devient :
```
submitted → approved → published → deleted | expired
submitted → rejected → (suppression auto)
submitted → challenged → (edit) → submitted (re-soumission)
```

### 1.2 Changement de comportement à la création

- `NewClassifiedAd` crée désormais en **`StatusSubmitted`** (au lieu de `StatusPublished`)
- `publishedAt` n'est plus renseigné à la création — il sera set lors de la transition `approved → published`
- L'event émis à la création devient `ClassifiedAdSubmittedEvent` (remplace `ClassifiedAdPublishedEvent` à la création)

### 1.3 Nouvelles transitions sur l'agrégat ClassifiedAd

| Méthode | De | Vers | Déclencheur |
|---------|-----|------|-------------|
| `Approve()` | submitted | approved | Consumer qui écoute l'event public `ClassifiedAdApproved` |
| `Publish()` | approved | published | Consumer qui écoute l'event interne `ClassifiedAdApproved` (chaîné immédiatement) |
| `Reject()` | submitted | rejected | Consumer qui écoute l'event public `ClassifiedAdRejected` |
| `Challenge()` | submitted | challenged | Consumer qui écoute l'event public `ClassifiedAdChallenged` |
| `Edit(...)` | challenged | submitted | Seller modifie l'annonce (tous les champs modifiables) |

Note : après `Reject()`, l'annonce est automatiquement supprimée (transition `rejected → deleted`).

### 1.4 Nouvelle commande : EditClassifiedAd

- **Précondition** : l'annonce est en statut `challenged`
- **Champs modifiables** : tous (titre, description, prix, images, catégorie, location)
- **Effet** : repasse l'annonce en `submitted`, émet un `ClassifiedAdEditedEvent`
- **Authentification** : même mécanisme que Delete (email + password du seller)

### 1.5 Nouvelle commande : PublishClassifiedAd

- **Précondition** : l'annonce est en statut `approved`
- **Effet** : passe en `published`, set `publishedAt`, émet `ClassifiedAdPublishedEvent`
- **Déclencheur** : consumer interne qui réagit à la réception de l'event public `ClassifiedAdApproved`

### 1.6 Consumers à ajouter dans ClassifiedAd

| Consumer | Écoute (event public) | Action |
|----------|----------------------|--------|
| `ClassifiedAdApprovedConsumer` | `ClassifiedAdApproved` | Appelle `PublishClassifiedAdCommand` |
| `ClassifiedAdRejectedConsumer` | `ClassifiedAdRejected` | Appelle `DeleteClassifiedAdCommand` (raison: rejected) |
| `ClassifiedAdChallengedConsumer` | `ClassifiedAdChallenged` | Met à jour le statut, envoie un email au seller |

**Convention** : chaque consumer est défini dans son **propre fichier** (un fichier = un consumer), avec son fichier de test associé. Exemple : `consumer/classified_ad_approved_consumer.go`, `consumer/classified_ad_rejected_consumer.go`, `consumer/classified_ad_challenged_consumer.go`. Cette règle s'applique à tous les consumers du projet (ClassifiedAd et Moderation).

### 1.7 DeleteReason

Ajouter `"rejected"` aux raisons existantes (`sold`, `no_more_to_sell`, `edit`). Conserver `"edit"` (le workaround delete+repost coexiste avec le vrai Edit sur challenged).

---

## 2. Bounded Context Moderation (nouveau)

### 2.1 Agrégat : Moderator

| Champ | Type | Description |
|-------|------|-------------|
| `id` | `uuid.UUID` | Identifiant unique |
| `fullName` | `string` | Nom complet pour traçabilité |

Pas d'authentification dans ce scope — le modérateur est considéré comme déjà authentifié en amont.

### 2.2 Agrégat : ModerationTask

| Champ | Type | Description |
|-------|------|-------------|
| `id` | `uuid.UUID` | Identifiant unique |
| `createdAt` | `time.Time` | Date de création |
| `classifiedAdID` | `string` | ID de l'annonce concernée |
| `moderatorID` | `*uuid.UUID` | ID du modérateur (nil si non claimée) |
| `claimedAt` | `*time.Time` | Date de claim (nil si non claimée) |

**Cycle de vie** : Créée → Claimée → Complétée (supprimée)

**Règles métier** :
- Une task en queue partagée est visible par tous les modérateurs
- `Claim(moderatorID)` : lock exclusif. Si déjà claimée → erreur `ErrTaskAlreadyClaimed`
- `Complete()` : uniquement par le modérateur qui a claimé → erreur `ErrNotTaskOwner` sinon
- Pas de timeout sur le claim
- Supprimée physiquement après complete (l'historique est porté par ClassifiedAdHistory)
- Émet un `ModerationTaskCompletedEvent` à la suppression

**Création** :
- Déclenchée par la réception de l'event public `ClassifiedAdSubmitted` ou `ClassifiedAdEdited`
- Chaque soumission/re-soumission crée une **nouvelle** task (nouvel ID)

### 2.3 Agrégat : ClassifiedAdHistory

**Modèle** : Liste append-only d'entrées (event log). Le statut courant est dérivé de la dernière entrée.

| Champ (agrégat) | Type | Description |
|-----------------|------|-------------|
| `classifiedAdID` | `string` | ID de l'annonce |
| `entries` | `[]HistoryEntry` | Liste chronologique |

**HistoryEntry** :

| Champ | Type | Description |
|-------|------|-------------|
| `occurredAt` | `time.Time` | Timestamp de l'événement |
| `action` | `HistoryAction` | Type d'action (enum) |
| `moderatorID` | `*string` | ID du modérateur (si action de modération) |
| `reason` | `*string` | Raison (si reject/challenge) |
| `snapshot` | `*ClassifiedAdSnapshot` | Contenu de l'annonce (si submitted/edited) |

**HistoryAction** (enum) :
- `submitted` — annonce soumise/re-soumise
- `approved` — approuvée par un modérateur
- `rejected` — rejetée par un modérateur
- `challenged` — challengée par un modérateur
- `published` — publiée
- `deleted` — supprimée
- `expired` — expirée

**ClassifiedAdSnapshot** :

| Champ | Type |
|-------|------|
| `title` | `string` |
| `description` | `string` |
| `priceInCents` | `int64` |
| `imageURLs` | `[]string` |
| `category` | `string` |
| `zipCode` | `string` |
| `cityName` | `string` |
| `sellerEmail` | `string` |
| `sellerPseudo` | `string` |

**Alimenté par** : consumer qui écoute tous les events publics de ClassifiedAd et Moderation.

### 2.4 Commandes Moderation

| Commande | Input | Préconditions | Effet | Event émis |
|----------|-------|---------------|-------|-----------|
| `ClaimModerationTask` | taskID, moderatorID | Task existe, non claimée | Lock la task pour le modérateur | `ModerationTaskClaimedEvent` |
| `AcceptClassifiedAd` | taskID, moderatorID | Task claimée par CE modérateur | Supprime la task | `ClassifiedAdApprovedEvent` (interne) → event public |
| `RejectClassifiedAd` | taskID, moderatorID, reason | Task claimée par CE modérateur, reason valide | Supprime la task | `ClassifiedAdRejectedEvent` (interne) → event public |
| `ChallengeClassifiedAd` | taskID, moderatorID, reason | Task claimée par CE modérateur, reason valide | Supprime la task | `ClassifiedAdChallengedEvent` (interne) → event public |

### 2.5 Raisons prédéfinies

**Raisons de rejet** (RejectReason) :
- `inappropriate_content` — Contenu inapproprié
- `suspect_price` — Prix suspect
- `wrong_category` — Catégorie incorrecte

**Raisons de challenge** (ChallengeReason) :
- `price_to_verify` — Prix à vérifier
- `category_to_fix` — Catégorie à corriger

### 2.6 Queries Moderation

| Query | Input | Output | Description |
|-------|-------|--------|-------------|
| `ListModerationTasks` | (filtres optionnels) | `[]ModerationTaskListItem` | Toutes les tasks actives (en attente + claimées) |
| `GetModerationTaskDetail` | taskID | `ModerationTaskDetailView` | Task + ClassifiedAdHistory + dernier snapshot |

**ModerationTaskListItem** :
- `id` : ID de la task
- `classifiedAdTitle` : titre de l'annonce
- `createdAt` : date de création
- `status` : "pending" ou "claimed"
- `claimedBy` : nom du modérateur (si claimée)

**ModerationTaskDetailView** :
- La task elle-même
- L'historique complet (ClassifiedAdHistory)
- Le dernier snapshot de contenu de l'annonce

---

## 3. Système d'événements publics

### 3.1 Infrastructure

- **Bus public** : un deuxième `InMemoryBus` séparé du bus interne de chaque BC
- **Localisation** : `internal/shared/` pour les contrats (DTOs des events publics)
- **Pattern** : chaque BC a un "publisher" qui consomme ses events internes (domain events) et produit les events publics correspondants sur le bus public

### 3.2 Format des events publics

DTOs minimalistes définis dans `internal/shared/`, découplés des domain events internes. Ce sont des **contrats d'intégration** entre les BC.

### 3.3 Events publics — ClassifiedAd → Public

| Event | Payload |
|-------|---------|
| `ClassifiedAdSubmitted` | `ClassifiedAdID`, `Title`, `Description`, `PriceInCents`, `ImageURLs`, `Category`, `ZipCode`, `CityName`, `SellerEmail`, `SellerPseudo`, `OccurredAt` |
| `ClassifiedAdEdited` | idem `Submitted` |
| `ClassifiedAdPublished` | `ClassifiedAdID`, `OccurredAt` |
| `ClassifiedAdDeleted` | `ClassifiedAdID`, `Reason`, `OccurredAt` |
| `ClassifiedAdExpired` | `ClassifiedAdID`, `OccurredAt` |

### 3.4 Events publics — Moderation → Public

| Event | Payload |
|-------|---------|
| `ClassifiedAdApproved` | `ClassifiedAdID`, `ModeratorID`, `OccurredAt` |
| `ClassifiedAdRejected` | `ClassifiedAdID`, `ModeratorID`, `Reason`, `OccurredAt` |
| `ClassifiedAdChallenged` | `ClassifiedAdID`, `ModeratorID`, `Reason`, `OccurredAt` |

### 3.5 Qui consomme quoi

| Consumer (BC) | Écoute (event public) | Action |
|---------------|----------------------|--------|
| Moderation | `ClassifiedAdSubmitted` | Crée une ModerationTask + ajoute entrée + snapshot dans ClassifiedAdHistory |
| Moderation | `ClassifiedAdEdited` | Crée une ModerationTask + ajoute entrée + snapshot dans ClassifiedAdHistory |
| Moderation | `ClassifiedAdPublished` | Ajoute entrée "published" dans ClassifiedAdHistory |
| Moderation | `ClassifiedAdDeleted` | Ajoute entrée "deleted" dans ClassifiedAdHistory |
| Moderation | `ClassifiedAdExpired` | Ajoute entrée "expired" dans ClassifiedAdHistory |
| ClassifiedAd | `ClassifiedAdApproved` | Appelle `PublishClassifiedAdCommand` |
| ClassifiedAd | `ClassifiedAdRejected` | Appelle suppression auto (delete reason: rejected) |
| ClassifiedAd | `ClassifiedAdChallenged` | Met à jour statut + envoie email au seller |

**Rappel** : un fichier par consumer (cf. convention en 1.6) — pas de fichier fourre-tout regroupant plusieurs consumers.

---

## 4. Structure de dossiers cible

```
internal/
├── shared/
│   ├── public_events.go          # DTOs des events publics (contrats)
│   └── public_event_bus.go       # Interface du bus public (réutilise pkg/eventbus.Bus)
├── classified-ad/
│   ├── domain/
│   │   ├── status.go             # +submitted, approved, challenged, rejected
│   │   ├── classified_ad.go      # +Approve(), Publish(), Reject(), Challenge(), Edit()
│   │   ├── event.go              # +ClassifiedAdSubmittedEvent, ClassifiedAdEditedEvent
│   │   ├── delete_reason.go      # +rejected
│   │   └── ...
│   ├── application/
│   │   ├── command/
│   │   │   ├── submit_classified_ad.go   # modifié: émet Submitted au lieu de Published
│   │   │   ├── edit_classified_ad.go     # NOUVEAU
│   │   │   ├── publish_classified_ad.go  # NOUVEAU (approved → published)
│   │   │   └── ...
│   │   └── query/
│   │       └── ...
│   └── adapter/
│       └── driving/
│           ├── http/             # +PUT /classified-ads/{id} pour Edit
│           ├── consumer/         # un fichier par consumer :
│           │   ├── classified_ad_approved_consumer.go
│           │   ├── classified_ad_rejected_consumer.go
│           │   └── classified_ad_challenged_consumer.go
│           └── publisher/        # NOUVEAU: publie events internes → bus public
├── moderation/
│   ├── domain/
│   │   ├── moderator.go
│   │   ├── moderation_task.go
│   │   ├── classified_ad_history.go
│   │   ├── history_entry.go
│   │   ├── snapshot.go
│   │   ├── reject_reason.go
│   │   ├── challenge_reason.go
│   │   ├── repository.go         # ModerationTaskRepository, ModeratorRepository, ClassifiedAdHistoryRepository
│   │   ├── event.go
│   │   └── errors.go
│   ├── application/
│   │   ├── command/
│   │   │   ├── claim_moderation_task.go
│   │   │   ├── accept_classified_ad.go
│   │   │   ├── reject_classified_ad.go
│   │   │   └── challenge_classified_ad.go
│   │   └── query/
│   │       ├── list_moderation_tasks.go
│   │       └── get_moderation_task_detail.go
│   └── adapter/
│       ├── driven/
│       │   └── inmemory/         # Repos in-memory
│       └── driving/
│           ├── http/             # Endpoints modération
│           ├── consumer/         # un fichier par consumer :
│           │   ├── classified_ad_submitted_consumer.go
│           │   ├── classified_ad_edited_consumer.go
│           │   ├── classified_ad_published_consumer.go
│           │   ├── classified_ad_deleted_consumer.go
│           │   └── classified_ad_expired_consumer.go
│           └── publisher/        # Publie events internes → bus public
```

---

## 5. Endpoints HTTP cible

### ClassifiedAd (ajouts)

| Méthode | Route | Commande |
|---------|-------|----------|
| `PUT` | `/classified-ads/{id}` | `EditClassifiedAdCommand` |

### Moderation (nouveau)

| Méthode | Route | Commande/Query |
|---------|-------|----------------|
| `GET` | `/moderation/tasks` | `ListModerationTasks` |
| `GET` | `/moderation/tasks/{id}` | `GetModerationTaskDetail` |
| `POST` | `/moderation/tasks/{id}/claim` | `ClaimModerationTask` |
| `POST` | `/moderation/tasks/{id}/accept` | `AcceptClassifiedAd` |
| `POST` | `/moderation/tasks/{id}/reject` | `RejectClassifiedAd` |
| `POST` | `/moderation/tasks/{id}/challenge` | `ChallengeClassifiedAd` |

---

## 6. Tests end-to-end

Les tests e2e existants (`e2e/classified_ad_lifecycle_test.go`) doivent être **augmentés** pour couvrir les différents parcours possibles entre Moderation et ClassifiedAd, de bout en bout (HTTP → commandes → events internes → publishers → bus public → consumers → HTTP), avec le wiring complet des deux BC (bus interne de chaque BC + bus public + publishers + consumers).

### 6.1 Parcours à couvrir

| Parcours | Étapes |
|----------|--------|
| **Happy path : approbation** | Soumission annonce → task créée en modération → claim → accept → annonce `approved` puis `published` (chaîné) → visible dans les queries publiques |
| **Rejet** | Soumission → claim → reject (avec reason) → annonce supprimée automatiquement (raison `rejected`) → non visible dans les queries publiques |
| **Challenge puis correction** | Soumission → claim → challenge (avec reason) → annonce `challenged` → seller édite (`PUT`) → annonce repasse `submitted` → **nouvelle** task créée → claim → accept → publiée |
| **Challenge puis rejet** | Soumission → challenge → edit → re-soumission → claim → reject → annonce supprimée |
| **Challenges multiples** | Soumission → challenge → edit → challenge à nouveau → edit → accept → publiée (l'historique contient toutes les entrées) |
| **Concurrence sur le claim** | Deux modérateurs tentent de claim la même task → un seul réussit, l'autre reçoit `ErrTaskAlreadyClaimed` (HTTP 409) |
| **Complete par un non-owner** | Un modérateur tente accept/reject/challenge une task claimée par un autre → `ErrNotTaskOwner` (erreur HTTP) |
| **Historique complet** | Après un parcours complet, `GetModerationTaskDetail` (ou l'historique) reflète toutes les entrées dans l'ordre : submitted, challenged, submitted, approved, published… avec snapshots et moderatorID |
| **Expiration après publication** | Annonce approuvée + publiée → horloge avancée de 90 jours → expirée → entrée `expired` dans l'historique |

### 6.2 Assertions attendues

Pour chaque parcours, vérifier :
- le **statut** de l'annonce côté ClassifiedAd (via les queries/HTTP publics) ;
- l'**état des tasks** côté Moderation (`GET /moderation/tasks`) — créées, claimées, supprimées après complete ;
- le contenu du **ClassifiedAdHistory** (entrées, ordre, snapshots, moderatorID, reasons) ;
- la **visibilité publique** : une annonce non publiée (`submitted`, `approved`, `challenged`, `rejected`) n'apparaît pas dans les listings publics.

---

## 7. Ordre d'implémentation suggéré

1. **`internal/shared/`** — Events publics (DTOs) + bus public
2. **Statuts ClassifiedAd** — Ajouter submitted/approved/challenged/rejected + modifier NewClassifiedAd
3. **Transitions ClassifiedAd** — Approve(), Publish(), Reject(), Challenge(), Edit()
4. **Commandes ClassifiedAd** — EditClassifiedAdCommand, PublishClassifiedAdCommand
5. **Publisher ClassifiedAd** — Publie les events internes vers le bus public
6. **Scaffold Moderation** — Structure de dossiers
7. **Domain Moderation** — Moderator, ModerationTask, ClassifiedAdHistory
8. **Commandes Moderation** — Claim, Accept, Reject, Challenge
9. **Queries Moderation** — List, GetDetail
10. **Consumers Moderation** — Écoute events publics ClassifiedAd
11. **Publisher Moderation** — Publie events internes vers le bus public
12. **Consumers ClassifiedAd** — Écoute events publics Moderation (Approved, Rejected, Challenged)
13. **HTTP Moderation** — Handler + routes
14. **Wiring dans main.go** — Bus public, consumers, publishers
15. **Tests e2e** — Augmenter `e2e/` avec les parcours de la section 6

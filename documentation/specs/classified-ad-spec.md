# Spec — Bounded Context `ClassifiedAd` (itération 1)

## 1. Contexte

**DDD tactique + architecture hexagonale**.

### Périmètre de cette itération

**Inclus** — le cycle de vie de l'annonce :
1. **Dépôt** (`Submit ClassifiedAd`)
2. **Recherche / consultation** (`ClassifiedAd SEARCHED WITH PARAMETERS`, `ClassifiedAd DISPLAYED`)
3. **Offre** (`Make offer`, `Buyer offer made`, `Offer email sent`)
4. **Suppression** (`Delete ClassifiedAd`, 3 raisons)
5. **Expiration** (`90 days passed since published` → `Online ClassifiedAd expired`)

**Explicitement hors périmètre** :
- La persistance SQL (in-memory uniquement).

---

## 2. Décisions structurantes (récapitulatif des arbitrages)

| Sujet | Décision |
|---|---|
| Modération | Hors périmètre → l'annonce est **publiée immédiatement** au dépôt |
| Expiration | **90 jours après publication**, une seule règle |
| Déclenchement de l'expiration | **Ticker** dans `cmd/api/main.go` (intervalle 1h) appelant une commande dédiée |
| Auth vendeur | **email + password hashé** (bcrypt) — pas de compte utilisateur |
| Seller | **Entité interne à l'agrégat**, dupliquée par annonce, pas de repository Seller |
| Offre | **Event pur** : ne mute pas l'agrégat, aucune offre stockée |
| Suppression | **Soft delete avec raison** |
| Suppression répétée | **Idempotente**, aucun second event |
| Images | **URLs fournies au submit** (pas de commande Upload, pas de port de stockage) |
| Prix | **≥ 0** (0 = don), en **centimes**, EUR |
| Visibilité | `status` = source de vérité, `IsOnline()` dérivé ; détail d'une annonce hors-ligne → **404** |
| Recherche | catégorie + localisation + fourchette de prix + mots-clés, tri configurable + pagination |
| ID | **UUID généré dans le constructeur du domaine** |
| Emails | **Consumers** abonnés aux events + port `pkg/mailer.Mailer` |
| Persistance | **In-memory** uniquement |
| API | REST minimal, 5 endpoints |

---

## 3. Modèle du domaine

### 3.1 Agrégat racine `ClassifiedAd`

Package : `internal/classified-ad/domain`, fichier `classified_ad.go`.

| Champ | Type | Règles de validation |
|---|---|---|
| `id` | `uuid.UUID` | généré par `NewClassifiedAd()` |
| `title` | `string` | non vide, ≤ 100 caractères |
| `description` | `string` | non vide, ≤ 4000 caractères |
| `price` | `Price` (VO) | ≥ 0, en centimes, devise EUR |
| `seller` | `Seller` (entité interne) | voir 3.3 |
| `status` | `Status` (enum) | `Published` \| `Deleted` \| `Expired` |
| `imageURLs` | `[]string` | **0 à 10**, chaque URL non vide |
| `submissionDate` | `SubmissionDate` (VO) | date de dépôt |
| `publishedAt` | `time.Time` | égal à la date de dépôt dans cette itération (pas de modération) |
| `category` | `Category` (enum) | `immo` \| `auto` \| `consumer_goods` \| `holidays` |
| `location` | `Location` (entité interne) | voir 3.3 |
| `deletedAt` | `*time.Time` | renseigné au soft delete |
| `deleteReason` | `DeleteReason` | `sold` \| `no_more_to_sell` \| `edit` |
| `expiredAt` | `*time.Time` | renseigné à l'expiration |

### 3.2 Cycle de vie

```
NewClassifiedAd()
        │
        ▼
   [Published]  ──── Delete(email, password, reason) ───▶ [Deleted]   (soft, terminal)
        │                                                     ▲
        └──── Expire(now)  (publishedAt + 90j) ──▶ [Expired] ──┘
                                              (le vendeur peut encore supprimer)
```

- `Published` est l'**état initial** : le dépôt publie immédiatement.
- `IsOnline() bool` ⇔ `status == Published`. C'est ce prédicat qui pilote la visibilité
  en recherche et en consultation.
- `Expired` reste supprimable par le vendeur.
- `Deleted` est terminal.

### 3.3 Value Objects et entités internes

**`Price`** (VO) — `amountInCents int64`, devise EUR implicite.
`NewPrice(cents int64)` → erreur si `< 0`. `0` est **valide** (don).

**`Email`** (VO) — validation de format par regex simple, non vide, normalisé en minuscules.

**`Password`** (VO) — valeur en clair validée à la création (**min 8 caractères**) puis
**stockée hashée** (bcrypt) via un port `PasswordHasher` (voir 3.5).
Le VO ne doit jamais exposer le clair (pas de `String()` révélateur).

**`SubmissionDate`** (VO) — encapsule la date de dépôt.

**`Seller`** (entité interne) : `Email` (VO), `Pseudo` (non vide, ≤ 30 caractères),
`Password` (VO, hashé).

**`Location`** (entité interne) : `ZipCode` (**code postal FR : exactement 5 chiffres**),
`CityName` (non vide).

**`Category`** (enum string) : `immo`, `auto`, `consumer_goods`, `holidays`.
Toute autre valeur → `ErrInvalidCategory`.

**`DeleteReason`** (enum string) : `sold`, `no_more_to_sell`, `edit`
(les trois stickies `Delete ClassifiedAd` de l'event storming).

### 3.4 Comportements de l'agrégat

| Méthode | Règles |
|---|---|
| `NewClassifiedAd(...)` | valide tous les champs, génère l'UUID, `status = Published`, `publishedAt = submissionDate` |
| `IsOnline() bool` | `status == Published` |
| `Delete(email, password, reason, hasher, now)` | vérifie que l'email correspond au vendeur **et** que le password matche le hash → sinon `ErrInvalidCredentials`. Si déjà `Deleted` : **no-op idempotent** (le retour doit permettre à la commande de savoir qu'aucun event ne doit être émis). Sinon → `Deleted` + `deletedAt` + `deleteReason` |
| `Expire(now)` | no-op si non `Published` ; sinon si `now >= publishedAt + 90j` → `Expired` + `expiredAt` |
| `IsExpirable(now) bool` | `status == Published && now >= publishedAt + 90j` |
| `CanReceiveOffer() bool` | `IsOnline()` |

**Constante domaine** : `AdLifetime = 90 * 24 * time.Hour`.

### 3.6 Erreurs du domaine

`ErrEmptyTitle`, `ErrTitleTooLong`, `ErrEmptyDescription`, `ErrDescriptionTooLong`,
`ErrNegativePrice`, `ErrInvalidEmail`, `ErrEmptyPseudo`, `ErrPseudoTooLong`,
`ErrPasswordTooShort`, `ErrInvalidZipCode`, `ErrEmptyCityName`, `ErrInvalidCategory`,
`ErrInvalidDeleteReason`, `ErrTooManyImages`, `ErrEmptyImageURL`,
`ErrInvalidCredentials`, `ErrAdNotAvailable`, `ErrClassifiedAdNotFound`,
`ErrEmptyOfferMessage`, `ErrOfferMessageTooLong`, `ErrNegativeOfferAmount`.

### 3.7 Domain events (`domain/event.go`)

Quatre events, conformes à `eventbus.DomainEvent` (`EventType() string`) :

| Struct | `EventType()` | Payload |
|---|---|---|
| `ClassifiedAdPublishedEvent` | `"ClassifiedAdPublished"` | `AdID`, `Title`, `Category`, `SellerEmail`, `SellerPseudo`, `PublishedAt` |
| `BuyerOfferMadeEvent` | `"BuyerOfferMade"` | `AdID`, `AdTitle`, `SellerEmail`, `BuyerEmail`, `BuyerPseudo`, `Amount` (centimes), `Message`, `OccurredAt` |
| `ClassifiedAdDeletedEvent` | `"ClassifiedAdDeleted"` | `AdID`, `Reason`, `DeletedAt` |
| `ClassifiedAdExpiredEvent` | `"ClassifiedAdExpired"` | `AdID`, `SellerEmail`, `ExpiredAt` |

Constructeurs de confort `New{Event}FromClassifiedAd(ad *ClassifiedAd, ...)`.
**Les events sont émis APRÈS persistance réussie** (règle projet).

---

## 4. Couche application

### 4.1 `SubmitClassifiedAdCommand` — `application/command/submit_classified_ad.go`

Entrée : `Title, Description, PriceInCents, SellerEmail, SellerPseudo, SellerPassword,
ImageURLs []string, Category, ZipCode, CityName`.

Flux : hash du password (port `PasswordHasher`) → `NewClassifiedAd(...)` (date via `Clock`)
→ `repo.Save` → publish `ClassifiedAdPublished`.
Retour : l'ID de l'annonce créée.

### 4.2 `MakeOfferCommand` — `application/command/make_offer.go`

Entrée : `AdID, BuyerEmail, BuyerPseudo, AmountInCents, Message`.

Règles :
- annonce introuvable → `ErrClassifiedAdNotFound`
- annonce non online (Deleted/Expired) → `ErrAdNotAvailable`
- `AmountInCents >= 0`, **aucune contrainte relative au prix demandé** (le vendeur juge)
- `Message` **obligatoire**, non vide, ≤ 1000 caractères
- `BuyerEmail` valide

Effet : **aucune mutation ni sauvegarde de l'agrégat**, publish `BuyerOfferMade`.

### 4.3 `DeleteClassifiedAdCommand` — `application/command/delete_classified_ad.go`

Entrée : `AdID, Email, Password, Reason`.

Flux : `FindByID` → `ad.Delete(...)` → si déjà supprimée : **retour succès sans `Save`
ni event** ; sinon `repo.Save` → publish `ClassifiedAdDeleted`.
Credentials invalides → `ErrInvalidCredentials` (**aucun effet de bord**).

### 4.4 `ExpireOutdatedAdsCommand` — `application/command/expire_outdated_ads.go`

Sans paramètre. `repo.FindExpirable(clock.Now())` → pour chacune `ad.Expire(now)` →
`repo.Save` → publish `ClassifiedAdExpired`. Retourne le nombre d'annonces expirées.

### 4.5 `SearchClassifiedAdsQuery` — `application/query/search_classified_ads.go`

Prend les critères (cf. `SearchCriteria`), retourne `[]ClassifiedAdListItemView`
(vue de liste : `ID, Title, PriceInCents, Category, CityName, ZipCode, FirstImageURL,
SubmissionDate`). **Jamais d'entité domaine exposée.**

Défauts : `SortBy = date_desc`, `Limit = 20`, `Offset = 0`.

### 4.6 `GetClassifiedAdQuery` — `application/query/get_classified_ad.go`

Par ID. Retourne `ClassifiedAdView` (`ID, Title, Description, PriceInCents, Category,
SellerPseudo, ImageURLs, Location, SubmissionDate`) — **jamais l'email ni le password du
vendeur**. Si l'annonce n'est pas online → `ErrClassifiedAdNotFound` (⇒ 404).

---

## 5. Adapters

### 5.1 Driven — `adapter/driven/`

- `inmemory/classified_ad_repository.go` : map protégée par mutex, implémente les
  4 méthodes du port, filtrage/tri/pagination en mémoire.
- `bcrypt/password_hasher.go` : implémente `PasswordHasher` (`golang.org/x/crypto/bcrypt`,
  déjà dans `go.mod`).
- `clock/system_clock.go` : `Clock` réel ; un `FixedClock` de test côté tests.

### 5.2 Driving — HTTP `adapter/driving/http/`

REST minimal (`handler.go` + `dto.go`) :

| Verbe | Route | Description |
|---|---|---|
| `POST` | `/classified-ads` | dépôt → `201` + `{id}` |
| `GET` | `/classified-ads?category=&zip=&city=&minPrice=&maxPrice=&q=&sortBy=&limit=&offset=` | recherche |
| `GET` | `/classified-ads/{id}` | détail — `404` si non online |
| `POST` | `/classified-ads/{id}/offers` | offre → `202`/`201` |
| `DELETE` | `/classified-ads/{id}` | body `{email, password, reason}` → `204` |

Mapping des erreurs : validation domaine → `400` ; `ErrClassifiedAdNotFound` → `404` ;
`ErrInvalidCredentials` → `403` ; `ErrAdNotAvailable` → `409`.

### 5.3 Driving — Consumers `adapter/driving/consumer/`

| Consumer | Écoute | Action |
|---|---|---|
| `AdPublishedEmailConsumer` | `ClassifiedAdPublished` | mail de confirmation de dépôt au vendeur |
| `OfferEmailConsumer` | `BuyerOfferMade` | mail au vendeur avec pseudo/email acheteur, montant, message |

Les deux utilisent `pkg/mailer.Mailer` (`FakeMailer` en dev, `MailerSpy` en test).
Un consumer d'email d'expiration n'est **pas** demandé pour l'instant.

### 5.4 Wiring — `cmd/api/main.go`

Instancier : eventbus in-memory, repo in-memory, hasher bcrypt, clock système, mailer fake,
les 4 commandes et 2 queries, le handler HTTP + routes, l'abonnement des 2 consumers,
et un **ticker de 1 heure** (goroutine) appelant `ExpireOutdatedAdsCommand`.
Conserver l'endpoint `/health` existant.

---

## 6. Stratégie de test

- **Domaine** : tests unitaires purs (constructeur, chaque règle de validation, `Delete`
  avec bons/mauvais credentials, idempotence, `Expire` avant/après 90j).
- **Application** : repo in-memory + `pkg/eventbus/testing.EventCollector` + `FixedClock`
  + hasher fake. Vérifier systématiquement : valeur retournée, état persisté, event émis.
  Sur les cas d'erreur : **aucun effet de bord** (rien de sauvegardé, aucun event).
- **Consumers** : `pkg/mailer/testing.MailerSpy`.
- `testify` est disponible (`github.com/stretchr/testify`).

---

## 7. Conventions & rappels projet

- Module Go : `ddd-second-hand-marketplace`. Fichiers en `snake_case`.
- Règle de dépendance : `Adapter → Application → Domain`, jamais l'inverse. Domaine pur
  (pas d'import d'infra, pas de `net/http`, pas de SQL).
- Suivre le skill `.claude/skills/tactical-ddd` (patterns détaillés, exemples de code).
- **Mettre à jour `internal/classified-ad/agent.md`** une fois l'implémentation faite
  (agent `subdomain-scribe`).
- Commandes utiles : `go test ./...`, `go vet ./...`, `go fmt ./...`,
  `go build -o bin/api ./cmd/api`.

---

## 8. Valeurs par défaut fixées (modifiables sans nouvelle validation)

- Intervalle du ticker d'expiration : **1 heure**.
- Tri par défaut de la recherche : **date décroissante**.
- Pagination par défaut : `limit = 20`.
- Longueur max du message d'offre : **1000 caractères**.

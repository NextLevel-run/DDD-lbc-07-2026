# feature post classified ad

## Prompt :

Contexte : Nous voulons faire une application de petites annonces pour un atelier sur le DDD. On a deja fait l'event storming. 

Objective : Créer la feature pour pouvoir poster une annonce que nous appellerons classified ad.

Aide moi à définir les élements de cet aggregat avant de coder.


## Reponse :

🍀 Avant de définir l'agrégat, j'ai besoin de comprendre ce qui est sorti de votre event storming. Plutôt que de deviner, cadrons ensemble. Voici comment je structurerais la réflexion, avec les questions clés à trancher.

1. Le cycle de vie (le cœur de l'agrégat)

Une annonce a typiquement plusieurs états. Les vôtres ressemblent-ils à ça ?

Draft → PendingReview → Published → (Sold | Withdrawn | Expired)

Questions :
- Y a-t-il une étape de modération avant publication, ou une annonce est-elle publiée directement ?
- Le brouillon (Draft) existe-t-il, ou l'annonce naît-elle déjà « publiable » ?
- Que se passe-t-il quand l'objet est vendu — est-ce l'annonce qui gère ça, ou un autre bounded context (Order/Transaction) ?

2. L'identité et le propriétaire

- Quel est l'identifiant ? (ClassifiedAdId en UUID, je suppose)
- Qui poste ? On garde juste un SellerId (référence vers un autre BC), ou l'agrégat porte plus d'infos sur le vendeur ?

3. Les Value Objects candidats

D'après ce qu'on voit habituellement sur une marketplace :

┌─────────────┬─────────────────────┬───────────────────────────────┐
│  Candidat   │        Rôle         │   Règle métier à confirmer    │
├─────────────┼─────────────────────┼───────────────────────────────┤
│ Title       │ Titre de l'annonce  │ longueur min/max ?            │
├─────────────┼─────────────────────┼───────────────────────────────┤
│ Description │ Texte descriptif    │ obligatoire ? longueur ?      │
├─────────────┼─────────────────────┼───────────────────────────────┤
│ Price       │ Montant + devise    │ prix > 0 ? gratuit autorisé ? │
├─────────────┼─────────────────────┼───────────────────────────────┤
│ Category    │ Classification      │ liste fermée ou libre ?       │
├─────────────┼─────────────────────┼───────────────────────────────┤
│ Photos      │ Collection d'images │ min 1 ? max N ?               │
├─────────────┼─────────────────────┼───────────────────────────────┤
│ Location    │ Où est l'objet      │ obligatoire ?                 │
└─────────────┴─────────────────────┴───────────────────────────────┘

4. Les invariants (le plus important en DDD)

C'est ce qui justifie l'existence de l'agrégat. Par exemple :
- « On ne peut pas publier une annonce sans titre / sans prix / sans au moins une photo »
- « On ne peut pas modifier une annonce vendue »
- « Seul le vendeur peut retirer son annonce »

Quelles règles doivent être garanties à tout instant dans votre modèle ?

5. Les Domain Events (issus de l'event storming)

Vous les avez déjà — quels sont-ils exactement ? Probablement :
- ClassifiedAdPublished
- ClassifiedAdPriceChanged
- ClassifiedAdSold
- ClassifiedAdWithdrawn

---
Pour avancer efficacement, dites-moi surtout :
1. Les événements exacts que votre event storming a produits
2. S'il y a une modération dans le flux
3. Le scope de la feature « poster une annonce » : juste la création/publication, ou aussi les modifications ultérieures ?

Une fois qu'on a ça, on fige les invariants et je pourrai coder l'agrégat proprement avec le skill tactical-ddd.


## Prompt 2 : 

Il y a une moderation avant la publication. L'annonce est directement publiable. Lorsque l'objet est vendu, le vendeur doit supprimer manuellement l'annonce, il doit ajouter la raison de pourquoi il supprime l'annonce.
Il y a bien un uuid pour l'annonce. Le proprietaire n'a pas de compte, toutes les infos sont dans l'annonce (mail, pseudo et mot de passe). Titre et description Obligatoire, prix >= 0, classfication: liste fermée, photo 0à
n. location obligatoire. Le vendeur est le seul à pouvoir supprimer son annonce. Le mot de passe ne doit pas etre stocké en clair.

Domaine event : 
- a classified ad has been posted
- a classified ad has been deleted by seller (mais ca va trop loin, on verra plus tard)

## Réponse 2 :

───────────────┼──────────────────────────┤
│ price       │ Price (VO)          │ >= 0                     │
├─────────────┼─────────────────────┼──────────────────────────┤
│ category    │ Category (VO/enum)  │ liste fermée             │
├─────────────┼─────────────────────┼──────────────────────────┤
│ photos      │ []Photo             │ 0 à n                    │
├─────────────┼─────────────────────┼──────────────────────────┤
│ location    │ Location (VO)       │ obligatoire              │
├─────────────┼─────────────────────┼──────────────────────────┤
│ seller      │ SellerContact (VO)  │ voir ci-dessous          │
├─────────────┼─────────────────────┼──────────────────────────┤
│ status      │ Status (enum)       │ état à confirmer (pt. 2) │
└─────────────┴─────────────────────┴──────────────────────────┘

Le point délicat : SellerContact et le mot de passe

Le vendeur n'a pas de compte, tout est dans l'annonce :

SellerContact (Value Object)
├── Email          (VO — format validé)
├── Nickname       (VO — le pseudo)
└── HashedPassword (VO — ne contient JAMAIS le clair)

Le hashing est une préoccupation d'infrastructure, pas du domaine pur. Ma reco pour garder le domaine propre :
- Un port PasswordHasher (interface dans le domaine)
- Le handler de commande hashe le clair via ce port, puis construit HashedPassword avec le hash
- Le VO HashedPassword ne peut être construit qu'à partir d'un hash — impossible d'y mettre du clair par accident

Ainsi le mot de passe en clair ne franchit jamais la frontière du domaine.

Les 3 points à trancher

1. La devise du prix. « prix >= 0 » — mais Price = montant seul, ou montant + devise (Money avec EUR) ? Pour une marketplace je recommande Money{Amount, Currency}, quitte à figer EUR par défaut. Ton choix ?

2. L'état à la création. Tu dis « modération avant publication » ET « directement publiable ». Je l'interprète comme : pas de brouillon, mais après le post l'annonce est en attente de modération. Donc :

[post] → PendingReview → (modération) → Published → ...

L'annonce naît-elle en PendingReview ? Ou considères-tu qu'au scope actuel (juste poster), on la crée directement en Published et on ajoutera la modération après ?

3. Photo. Une photo = juste une URL/référence (string validée), ou tu veux d'autres attributs (légende, ordre d'affichage) ?

Dès que tu réponds à ces 3 points, je code l'agrégat avec le skill tactical-ddd.

## Prompt 3 : 

1. on est mono devise mais on peut avoir des prix au centime près.
2. L'annonce nait en PendingReview
3. Pour les photos, juste l'URL

## Réponse :

🍀 On a tout. Voici le modèle figé de l'agrégat ClassifiedAd.

Modèle final

Aggregate Root : ClassifiedAd

┌─────────────┬────────────────┬─────────────────────────────────────┐
│    Champ    │      Type      │              Invariant              │
├─────────────┼────────────────┼─────────────────────────────────────┤
│ id          │ ClassifiedAdId │ UUID                                │
├─────────────┼────────────────┼─────────────────────────────────────┤
│ title       │ Title          │ obligatoire, non vide               │
├─────────────┼────────────────┼─────────────────────────────────────┤
│ description │ Description    │ obligatoire, non vide               │
├─────────────┼────────────────┼─────────────────────────────────────┤
│ price       │ Price          │ montant en centimes (int64), >= 0   │
├─────────────┼────────────────┼─────────────────────────────────────┤
│ category    │ Category       │ liste fermée (enum)                 │
├─────────────┼────────────────┼─────────────────────────────────────┤
│ photos      │ []Photo        │ 0 à n, chaque photo = URL           │
├─────────────┼────────────────┼─────────────────────────────────────┤
│ location    │ Location       │ obligatoire                         │
├─────────────┼────────────────┼─────────────────────────────────────┤
│ seller      │ SellerContact  │ email + pseudo + mot de passe hashé │
├─────────────┼────────────────┼─────────────────────────────────────┤
│ status      │ Status         │ naît en PendingReview               │
└─────────────┴────────────────┴─────────────────────────────────────┘

Value Objects
- Price : stocke un int64 en centimes → pas de flottant, mono-devise implicite. Ex. 12,34 € = 1234. Invariant >= 0.
- Photo : une URL validée (non vide, format URL).
- SellerContact : { Email, Nickname, HashedPassword } — le clair ne rentre jamais dans le domaine.

Statuts (enum)
PendingReview → Published → (Sold | Withdrawn)   // seul PendingReview concerne cette feature

Port (domaine)
- PasswordHasher : interface pour hasher le mot de passe. Le handler l'appelle, puis construit HashedPassword avec le hash.

Domain Event
- ClassifiedAdPostedEvent (type string : "ClassifiedAdPosted") — émis après persistance.

Comportement du scope actuel
- Constructeur NewClassifiedAd(...) : valide tous les invariants, positionne status = PendingReview, enregistre l'événement ClassifiedAdPosted.

---
Deux micro-décisions d'implémentation que je propose de trancher par défaut (dis-moi si tu diverges) :

- Category : je pars sur un type énuméré Go avec un ensemble fixe (ex. Electronics, Furniture, Vehicles, Clothing, Other). Tu veux une liste précise ou ces exemples suffisent pour l'atelier ?
- Contenu de ClassifiedAdPostedEvent : je mets ClassifiedAdId, Title, Category, SellerEmail, PostedAt. Suffisant ?

Si ça te convient, je lance le codage avec le skill tactical-ddd (agrégat + VOs + port + event, en commençant par le domaine et ses tests). Tu veux que je démarre ?

## Prompt 4 : 

oui

## Réponse : 

word_hasher.go — port PasswordHasher
- classified_ad_test.go — VOs + invariants

Application (application/command/)
- post_classified_ad.go — flux : garde mot de passe non vide → hash via port → NewClassifiedAd (valide les invariants) → persist → émet l'event
- test complet, dont un cas dédié « le mot de passe n'est jamais stocké en clair »

Adapters driven
- adapter/driven/inmemory/repository.go
- adapter/driven/password/bcrypt_hasher.go (bcrypt)

Deux décisions à te signaler

1. Contenu de l'event : j'ai embarqué la référence à l'agrégat (event.ClassifiedAd *ClassifiedAd) plutôt que des champs plats. C'est la convention du codebase (cf. OrderPlacedEvent) et ça expose de toute façon Title, Category, SellerEmail, PostedAt. Si tu préfères des champs plats explicites, je le change.
2. go mod tidy a résolu golang.org/x/crypto de v0.43.0 → v0.54.0 (nécessaire pour bcrypt). C'est un changement dans go.mod/go.sum.

Prochaines étapes possibles

- Adapter driving HTTP (POST /classified-ads) + DTO + route
- Câblage dans cmd/api/main.go (repo + bcrypt hasher + event bus)
- Doc : mettre à jour internal/classified-ad/agent.md (actuellement vide) via l'agent subdomain-scribe

Tu veux que j'enchaîne sur le HTTP + le câblage, ou tu préfères d'abord relire le domaine ?
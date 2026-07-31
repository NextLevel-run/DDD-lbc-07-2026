# Objectif

Mettre en place le bounded context moderation et le lier avec ClassifiedAd

## Context
- On a mappé le métier dans un event storming (dans event-storming.md et event-storming.jpg)
- On a déjà une première version d'annonce fonctionnelle sans modération

### Il faut
- Créer le bounded context moderation
- Créer les Agregats:
  - **Moderator**: uuid, login, password, fullname
  - **ModerationTask**: uuid, creationtime, classifiedAdId, ModeratorId
    - Assumé quand la tâche est claimée par un modérateur
    - Méthodes : claim (lock la tâche pour un modérateur), complete (supprime la task de la queue)
  - **ClassifiedAdHistory**: Audit trail des statuts de la classified ad et des actions de modération
    - Accès facile au statut courant de la classifiedAd
    - Actions possibles: accept, reject (avec reason), challenge (avec reason pour le seller)
    - Applicable quand la classified ad est submited ou edited
- Créer un système d'événements publics partagés entre les différents Bounded Contexts
  - Utiliser un folder `internal/shared` pour partager les formats des événements publics
    - avoir une infra dédié pour le bus public
  - Produire les événements publics en consommant les événements privés
  - Tous les events de changement de statut de l'annonce sont public et consomme par Moderation pour pouvoir les afficher dans ClassifiedAdHistory
  - Le events de Moderation (Approved, Rejected et Challenged) en event public afin que le Bounded Context ClassifiedAd les consomme
- Ajouter un état "submitted" sur l'annonce et l'event lié
- **Statuts**: Approved, Rejected, Challenged
  - Rejected → delete
  - Challenged → edit
  - Approved → Publish

Interview moi jusqu'à ce qu'il n'y ait plus d'ambiguité fonctionnelle.

Objectif : créer un bounded context ClassifiedAd
Contexte :
* on a mappé le métier dans un event storming (dans event-storming.md et event-storming.jpg)
* on va implémenter une première partie de l'event storming : celle qui concerne le cycle de vie de l'annonce (dépot, recherche, offre, suppression)
* voilà l'agrégat qu'on a identifié pour ce bounded context : 
  * ClassifiedAd : id (uuid), title, description, prix (value object), seller (entity, email (value object), pseudo), password (VO utilisé pour le delete), status, images url, submission date (VO), category (enum: immo, auto, bien de consommation, vacances), location (entity zip code et city name)
* Pas d'agrégat offer, c'est juste un event

Interview moi jusqu'à ce qu'il n'y ait plus d'ambiguité fonctionnelle.
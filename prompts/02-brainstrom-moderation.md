
Objectif
Mettre en place le bounded context moderation et lie avec ClassifiedAd

context
- on a mappé le métier dans un event storming( dans event-storming.md et event-storming.jpg)
- on a deja une premiere version d'annonce fonctionnelle sans moderation
  il faut 
  - creer le bounded context moderation
  - creer les Aggregat Moderator, ModerationTask et ClassifiedAdHistory
    - Moderator(uuid, login, password, et fullname
    - ModerationTask(uuid, creationtime, classifiedAdId, ModeratorId(assume quand la tache est claim par un moderateur; Methodes : claim qui lock la tache pour un moderateur et complete qui supprime la task de la queue)
    - ClassifiedAd history: Audit trail des status de la classified ad et des actions de moderation
      - On doit pouvoir acceder facilement le status en cours de la classifiedAd   
      - On doit pouvoir accept, reject (avec reason pour des question de statistique interne) ou challenge (avec reason pour indiquer au seller la raison de son challenge) la classifiedAd quand elle est submited ou edited
-Ajouter un etat submited sur l annonce
-generer une moderation task
-avoir une moderation history
-envoyer des events public quand une annonce est rdy to mode 
- Approve, Rejected, Challenged comme status
- Rejected = delete
- Challeged = edit
- Approved = Publish
- Ajouter edit dans cad
- cad history 
- Ajouter d un life cycle expired, archived, rejected, challenged, approved, edited
- Ajouter une recherche dans histo
- valeur email, counter view, counter offer, reason

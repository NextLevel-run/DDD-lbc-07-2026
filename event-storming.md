# Event Storming — ClassifiedAd Marketplace

> 72 sticky notes total: 40 domain events, 12 command stickies (10 distinct commands),
> 7 actor stickies, 4 external system stickies, 9 view/read model stickies.

## Legend (color code used on the board)

| Color | Type |
|---|---|
| 🟠 Orange | Domain Event |
| 🔵 Dark blue | Command |
| 🟡 Light yellow | Actor |
| 🩷 Pink | External System |
| 🟢 Dark green | View / Read Model |

## Actors

- Seller
- Moderator
- Potential Buyer
- Reporter

## External Systems

- Email
- Ad Server

## Views / Read Models

- Homepage view
- Deposit form view
- Moderation list view
- Moderation ClassifiedAd view
- ClassifiedAd Edit view
- ClassifiedAd list view
- ClassifiedAd view
- Offer form view
- ClassifiedAd delete view

## Commands (10 distinct)

- Upload image
- Submit ClassifiedAd
- Select ClassifiedAd for moderation
- Approve ClassifiedAd
- Reject ClassifiedAd
- Challenge ClassifiedAd
- Edit ClassifiedAd
- Report ClassifiedAd
- Make offer
- Delete ClassifiedAd *(appears 3x on the board with different triggering reasons)*

## Domain Events (40)

- HOMEPAGE DISPLAYED
- DEPOSIT BUTTON CLICKED
- DEPOSIT CATEGORY SELECTED
- IMAGES UPLOADED
- ClassifiedAd Submitted
- ClassifiedAd DEP. EMAIL SENT
- ClassifiedAd QUEUED FOR MOD.
- ClassifiedAd SELECTED BY MOD
- ClassifiedAd APPROVED
- ClassifiedAd REJECTED
- ClassifiedAd CHALLENGED
- ClassifiedAd Challenge email sent
- ClassifiedAd published email sent
- ClassifiedAd published
- ClassifiedAd OFFLINE
- ClassifiedAd EDITED
- ClassifiedAd REPORTED ONCE
- ClassifiedAd REPORTED TWICE
- ClassifiedAd PUSHED BACK TO MOD
- AD SEARCH (MAP) REGION CLICKED
- ClassifiedAd SEARCHED WITH PARAMETERS
- ClassifiedAd LISTED
- Ad reached
- Ad displayed
- Ad clicked
- ClassifiedAd DISPLAYED
- BUY CTA CLICKED
- Buyer offer made
- Offer email sent
- SELLER REPLIED
- OFFER ACCEPTED
- Transaction occured
- ClassifiedAd Deleted for sell
- ClassifiedAd Deleted for no more to sell
- ClassifiedAd Deleted for edit
- 60 days passed since submit
- 90 days passed since published
- Online ClassifiedAd expired
- ClassifiedAd expired
- Ad paidout

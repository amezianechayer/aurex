# Guide d'intégration — proposer la Murabaha dans votre application

Ce guide s'adresse au développeur d'une fintech, néobanque ou plateforme de
financement qui veut proposer des contrats Murabaha conformes AAOIFI à ses
clients, au-dessus de l'API corren.

## Le modèle en une phrase

Votre application gère l'expérience client (KYC, écrans, prélèvements PSP) ;
corren gère **le contrat, la comptabilité et la conformité Sharia** — et il
est impossible d'exécuter une opération non conforme via l'API, même par bug.

## Répartition des outils

| Besoin | Outil |
|---|---|
| Cycle de vie Murabaha (création → règlement) | **API Contrats** (`/contracts`) |
| Mouvements d'argent libres (dépôts, wallets, commissions, remboursements) | **FaRl** (`POST /:ledger/script`) ou `POST /:ledger/transactions` |
| Back-office humain | **Horizon** (ou vos écrans sur la même API) |

Le contrat n'est *pas* un script FaRl : le moteur génère lui-même les
écritures (qabd, profit différé FAS 28, pénalités → charité). FaRl sert au
cash-flow autour du contrat.

## 0. Authentification

Demandez deux clés API à l'opérateur de la plateforme (ou générez-les avec
`corren auth init` si vous opérez vous-même) :

- une clé `operator` scopée à votre ledger — pour les écritures ;
- une clé `readonly` — pour vos dashboards et webhooks de lecture.

Toutes les requêtes : `Authorization: Bearer crn_…`. Une clé `readonly` qui
tente une transition reçoit 403 : branchez vos services de lecture dessus
sans risque.

## 1. Onboarding et dépôts (FaRl)

Les comptes ledger naissent au premier posting — pas d'étape de création.
Quand votre PSP confirme un dépôt client :

```bash
curl -X POST $API/monledger/script \
  -H "Authorization: Bearer $OPERATOR_KEY" \
  -d '{
    "plain": "var $c: account\ntransfer [DZD.2 5000000] (\n  from @world\n  to   $c\n)",
    "vars": {"c": "@client:anis"}
  }'
```

Provisionnez aussi votre trésorerie (`@bank:treasury`) : `acquire` vérifie
qu'elle couvre le coût du bien avant de payer le fournisseur.

## 2. Créer le contrat

```bash
curl -X POST $API/monledger/contracts \
  -H "Authorization: Bearer $OPERATOR_KEY" \
  -d '{
    "type": "murabaha",
    "params": {
      "asset_code": "VHCL884",
      "cost":   {"asset": "DZD.2", "amount": 150000000},
      "markup": {"asset": "DZD.2", "amount": 15000000},
      "client": "@client:anis",
      "supplier": "@supplier:toyota_alger",
      "installments": 24,
      "first_due": "2026-08-01T00:00:00Z",
      "period_days": 30
    }
  }'
```

La réponse contient l'échéancier complet, calculé au centime (int64, jamais
de float ; le reste de la division va sur la dernière échéance). Affichez-le
au client **avant** signature : c'est une exigence de transparence du
contrat Murabaha.

Règles à connaître :
- `markup` est un **montant fixe**, jamais un taux ;
- `cost` et `markup` sont immuables après création (aucun endpoint ne les
  modifie) ;
- l'`id` est optionnel (généré sinon) ; minuscules/chiffres/underscore.

## 3. Brancher le cycle de vie sur vos événements métier

```
PROMISE ──acquire──▶ ACQUIRED ──sell──▶ SOLD ──pay_installment ×N──▶ SETTLED
   │                    │                 ├──early_settle──▶ SETTLED
   └──cancel──▶ CANCELLED◀───cancel       └──late_penalty (charité)
```

| Événement chez vous | Appel corren |
|---|---|
| Achat du bien validé par votre back-office | `POST /contracts/:id/transitions/acquire` |
| Client signe la réception du bien | `POST /contracts/:id/transitions/sell` |
| Prélèvement mensuel confirmé par votre PSP | créditer le wallet (FaRl) puis `POST …/pay_installment` `{"seq": n}` |
| Demande de remboursement anticipé | `POST …/early_settle` `{"rebate": r}` (défaut : remise totale du profit non couru) |
| Retard (le scheduler a marqué `overdue`) | décision explicite → `POST …/late_penalty` `{"seq": n, "amount": p, "destination": "@charity:pool"}` |
| Renonciation avant livraison | `POST …/cancel` |

Réponse type : `{"contract_id", "transition", "new_state", "tx_id"}`.

### Erreurs à gérer

| Code | HTTP | Action côté app |
|---|---|---|
| `ERR_DUPLICATE` | 409 | la transition a déjà été exécutée — **retry sans danger**, traitez comme un succès |
| `ERR_PRECONDITION` | 422 | solde insuffisant (client ou trésorerie) — re-tentez après provisionnement |
| `ERR_INVALID_TRANSITION` | 409 | bug de séquencement chez vous — loggez, n'insistez pas |
| `ERR_SHARIA_VIOLATION` | 422 | opération interdite (avec `standard_ref` AAOIFI) — ne contournez jamais, le refus est journalisé |
| `ERR_INVALID_PARAMS` | 400 | corrigez le payload |

L'idempotence par référence unique signifie que vous pouvez mettre chaque
transition dans une file avec retry automatique : un double envoi ne
produira jamais une double écriture.

## 4. Ce que vous exposez à vos clients finaux

- **Échéancier** : `GET /contracts/:id/schedule` (statuts paid/pending/
  overdue/settled_early, tx de paiement).
- **Preuve de conformité** : `GET /contracts/:id/audit?verify=true` —
  chaîne d'audit hashée, chaque événement référencé au standard AAOIFI,
  `chain_valid` recalculé en direct. Affichable tel quel, et c'est la
  matière première du rapport annuel pour un Sharia board.
- **Soldes** : `GET /accounts/@client:anis` (balances, volumes).

## 5. Surveillance des retards

Le scheduler de corren marque `overdue` les échéances `pending` au-delà du
délai de grâce (`sharia.grace_days`, défaut 7 jours) — il journalise mais ne
déclenche **jamais** de pénalité automatiquement : la pénalité reste une
décision explicite, et sa seule destination possible est un compte
`@charity:*` (AAOIFI SS-3). Pour vos relances, sondez
`GET /contracts/:id/schedule` ou requêtez vos contrats `SOLD`.

## Checklist avant pilote réel

- [ ] `auth.enabled: true` + clés scopées par ledger
- [ ] TLS devant l'API (reverse proxy)
- [ ] Postgres en production (`storage.driver: postgres`) — testé par la
      suite d'intégration (`CORREN_TEST_PG_CONN`)
- [ ] Validation du produit par votre Sharia advisor (les références sont au
      niveau document : AAOIFI SS-8, SS-3, FAS 28)
- [ ] Cadre réglementaire : le financement est une activité licenciée dans
      la plupart des juridictions — adossez-vous à une institution agréée
      si vous ne l'êtes pas

## Hors périmètre v1

Ijarah/Mudarabah/autres contrats, calendrier hijri, restructuration de
dette, webhooks sortants (sondez les endpoints en attendant). Voir
`sharia/README.md` pour les invariants et le mapping standard → test.

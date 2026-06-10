# Design — Durcissement pilote client (auth, Postgres, Horizon UI, observabilité)

**Date** : 2026-06-10
**Statut** : validé par le user (approche A)
**Base** : branche `feature/sharia-murabaha-v1` (module Murabaha v1 complet)
**Objectif** : rendre corren + Horizon utilisables par un pilote client réel :
une fintech intègre l'API avec des clés à rôles, des opérateurs humains
supervisent et agissent via Horizon avec login.

Ordre de livraison : Phase 1 → 2 → 3 → 4, chaque phase testée (TDD) et
committée indépendamment. Branches : `feature/pilot-hardening` (corren,
empilée sur `feature/sharia-murabaha-v1`) et `feature/contracts-ui`
(repo horizon). Jamais de commit sur `main`.

---

## Phase 1 — Authentification & rôles (repo corren)

### Stockage

Les bases corren sont par-ledger (un fichier SQLite par ledger / un schéma
Postgres par ledger). L'authentification est globale → base dédiée
`corren_auth` (fichier `corren_auth.db` en SQLite, schéma `corren_auth` en
Postgres), gérée par un nouveau package `auth/` avec son propre store
(même style : go-sqlbuilder, migrations embarquées `--statement`).

Tables :

- `auth_users` : `id`, `username` (unique), `password_hash` (bcrypt),
  `role`, `created_at`, `disabled_at` (nullable)
- `auth_keys` : `id`, `key_hash` (sha256 du token), `label`, `role`,
  `ledgers` (CSV de noms ou `*`), `created_at`, `revoked_at` (nullable)
- `auth_sessions` : `id`, `user_id`, `token_hash` (sha256), `expires_at`,
  `created_at`, `revoked_at` (nullable)

Aucun secret en clair en base. Tokens opaques générés aléatoirement
(32 octets, préfixe `crn_` pour les clés API, `crs_` pour les sessions).
Pas de lib JWT : sessions révocables côté serveur, expiration configurable
(`auth.session_ttl`, défaut 12h). Seule dépendance :
`golang.org/x/crypto/bcrypt` (déjà présente dans go.sum en indirecte).

### Rôles et application

| Rôle | Droits |
|---|---|
| `admin` | tout + gestion users/clés (`/auth/*` admin) |
| `operator` | POST contrats, transitions, transactions, scripts + tous les GET |
| `readonly` | GET uniquement |

Application à deux niveaux :
1. middleware par méthode/route (GET → readonly+, POST → operator+,
   `/auth/admin/*` → admin) ;
2. scope ledger : une clé limitée à `["demo"]` ne peut rien faire sur un
   autre ledger (middleware compare au paramètre `:ledger`).

### Endpoints

- `POST /auth/login` `{username,password}` → `{token, expires_at, role}`
- `POST /auth/logout` (révoque la session)
- `GET  /auth/me` → rôle + identité (utilisé par Horizon)
- Admin : `POST/GET/DELETE /auth/admin/keys`, `POST/GET /auth/admin/users`
  (la clé API n'est affichée qu'à la création, jamais relisible)

Accès : header `Authorization: Bearer <token>` (clé API ou session,
distinguées par préfixe).

### Bootstrap & compatibilité

- `corren auth init` : crée le premier user admin (mot de passe demandé ou
  généré) + une clé admin, affichées une seule fois.
- Flag `auth.enabled` (défaut `false` + warning au démarrage). Désactivé =
  comportement actuel inchangé (démos, dev). Le pilote l'active.
- L'ancien `server.http.basic_auth` reste fonctionnel mais déprécié (warning).

### Tests (TDD)

Store auth (round-trip, révocation, expiration), middleware (matrice
rôle × méthode × route, scope ledger, token inconnu/expiré/révoqué → 401/403),
httptest bout-en-bout : login → action → logout → 401. Un test vérifie
qu'une clé `readonly` reçoit 403 sur `POST .../transitions/sell`.

---

## Phase 2 — Postgres prouvé

- Service `postgres` ajouté à `docker-compose.yml` (image officielle,
  healthcheck, port non conflictuel).
- Tests d'intégration des scénarios A–G + store sharia exécutés contre
  Postgres, activés par `CORREN_TEST_PG_CONN` (absent → skip, `go test ./...`
  reste vert sans Postgres). Factorisation des helpers de test existants
  pour être paramétrés par driver.
- Correction des bugs découverts dans `storage/postgres/sharia.go`
  (code jamais exécuté à ce jour) et dans la migration v002 Postgres.

Critère : scénarios A–G verts sur Postgres dans le conteneur compose.

---

## Phase 3 — Frontend Horizon (repo horizon, branche `feature/contracts-ui`)

Stack inchangée : React 16 + webpack + styled-components + react-table +
axios (cohérence avec les pages Accounts/Transactions existantes ; pas de
réécriture).

### Pages

1. **Login** (`/login`) : username/password → `POST /auth/login`, token en
   mémoire + localStorage, intercepteur axios (header Bearer + redirect 401).
   Si l'API répond que l'auth est désactivée, accès direct (mode dev).
2. **Contrats** (`/ledgers/:ledger/contracts`) : tableau (id, type, état
   avec badge coloré, coût, marge, progression x/N échéances payées),
   filtre par état, bouton « Nouveau contrat ».
3. **Détail contrat** (`/ledgers/:ledger/contracts/:id`) :
   - en-tête : état + paramètres (coût, marge, client, fournisseur) ;
   - timeline du cycle de vie (PROMISE → … → SETTLED) construite depuis
     l'audit trail ;
   - échéancier : tableau seq/date/montant/principal/profit/statut avec
     code couleur (payé/en attente/en retard/réglé par anticipation) ;
   - **audit trail** : liste des événements (allowed en vert, denied en
     rouge avec `standard_ref`), badge `chain_valid` issu de
     `GET …/audit?verify=true` — l'argument commercial central ;
   - actions : uniquement les transitions légales depuis l'état courant,
     modal de confirmation avec champs requis (seq, rebate, pénalité),
     erreurs API affichées avec leur `standard_ref`.
4. **Création** (`/ledgers/:ledger/contracts/new`) : formulaire des params
   avec validation client, soumission → redirection vers le détail
   (l'échéancier affiché est celui renvoyé par l'API — pas de recalcul JS).

Rôle `readonly` (via `GET /auth/me`) : boutons d'action masqués.

### Tests

Le repo horizon n'a pas d'infra de test : tests manuels structurés via
checklist + vérification des golden paths au navigateur (browse/QA),
captures avant/après. Pas d'introduction de Jest dans cette phase (YAGNI
pour un pilote ; à reconsidérer en v2).

---

## Phase 4 — Observabilité pilote

- `GET /healthz` : ping de la base auth (toujours présente) + version
  → 200/503 (sans auth, pour les load-balancers).
- Logs structurés (JSON une ligne) sur chaque transition et chaque refus :
  contract_id, transition, decision, standard_ref, tx_id, durée.
- Extension de `GET /:ledger/stats` : contrats par état, refus journalisés,
  échéances en retard.
- Pas de Prometheus en pilote (nouvelle dépendance pour peu de valeur ;
  reconsidérer après le pilote).

---

## Hors périmètre (rappel)

Calendrier hijri, autres contrats (Ijarah…), restructuration de dette,
multi-madhab, certification Sharia humaine, KYC/AML, Prometheus, refonte
de la stack Horizon.

## Risques connus

- `storage/postgres/sharia.go` jamais exécuté → phase 2 peut révéler des
  surprises (prévu, c'est son but).
- React 16/webpack 4 sont datés : on assume la dette pour le pilote.
- Le mode `auth.enabled=false` doit rester strictement identique au
  comportement actuel (testé en non-régression).

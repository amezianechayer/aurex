# Design — extract `sharia/` into a standalone `corren-sharia` Go module

**Date** : 2026-06-17
**Statut** : modèle validé par le user
**Décisions** : (1) module Go importé par corren — dépendance `corren → corren-sharia → corren-vm` ; (2) même PostgreSQL que corren, mais tables + migrations **propres** à corren-sharia.

## 1. But

Sortir toute la **logique Sharia** (moteur multi-contrats, FSM murabaha/ijarah, échéanciers, audit hash-chainé AAOIFI, validation, gate) de corren vers un module autonome `github.com/amezianechayer/corren-sharia` :
- son propre `go.mod` ;
- son propre storage (sqlite + postgres) + ses migrations ;
- ses handlers REST (mêmes endpoints) ;
- consomme `corren-vm` ; **n'importe jamais `corren`** (sinon cycle de modules).

corren ne garde que le **câblage** : fournir un ledger (via un adaptateur) + monter les routes. Aucune *logique* Sharia ne reste.

## 2. Contrainte structurante : les types `core`

Le code sharia de prod n'importe aujourd'hui que `corren/core` (Posting/Transaction/Account/Metadata/Script/WORLD) et `corren/config` (allowed_assets). Comme `corren → corren-sharia`, le module **ne peut pas** importer `corren/core` (cycle de modules). Donc :

- corren-sharia définit son **propre package `chain`** : `Posting`, `Transaction`, `Account`, `Metadata`, `Script`, const `WORLD` — structurellement identiques à `corren/core`.
- `LedgerPort` (déjà dans sharia) s'exprime en types `chain`.
- corren fournit un **adaptateur** `shariaLedger{ *ledger.Ledger }` qui satisfait `corren-sharia.LedgerPort` en convertissant `chain.Transaction ↔ core.Transaction` (copie de champs — types identiques).
- `corren/config` → remplacé par une `Config{ AllowedAssets []string }` injectée à la construction du moteur.

## 3. Découpage des responsabilités

| Aujourd'hui (corren) | Demain |
|---|---|
| `sharia/*.go` (9 fichiers prod) | → `corren-sharia/` (engine, kinds, fsm, types, audit, schedule, errors) |
| `storage/{sqlite,postgres}/sharia.go` | → `corren-sharia/storage/{sqlite,postgres}` (impl `ShariaStore`) |
| migrations v002 (+ part v003) | → migrations **propres** de corren-sharia (mêmes tables, run par corren-sharia sur le même DB) |
| `api/actions/contract_controller.go` | → handlers REST dans `corren-sharia/api` ; corren monte les routes |
| `storage.Store` embarque `ShariaStore` | corren-sharia gère SON storage ; `storage.Store` **ne** l'embarque **plus** |
| `flows` importe `sharia` | `flows` importe `corren-sharia` (ShariaPort en types corren-sharia) |
| `scheduler` RunOnce (overdue) | utilise corren-sharia |

**Migrations sur le même DB :** corren-sharia expose un `Migrate(db)` (ou un `Store.Initialize()`) qui crée SES tables (`sharia_contracts`, `installments`, `sharia_audit`) de façon idempotente (`CREATE TABLE IF NOT EXISTS`). corren l'appelle à l'init du ledger, après ses propres migrations. Deux runners, un DB, pas de collision (tables disjointes).

## 4. Endpoints préservés (à l'identique)

`POST /:ledger/contracts`, `GET /:ledger/contracts`, `GET /:ledger/contracts/:id`,
`POST /:ledger/contracts/:id/transitions/:name`, `GET /:ledger/contracts/:id/schedule`,
`GET /:ledger/contracts/:id/audit`. Mêmes formats de réponse + erreurs typées (`sharia.Error` → HTTPStatus, dont le mapping guard-deny→422 déjà en place côté corren reste au niveau wiring).

## 5. Plan en 2 phases (chaque phase garde corren vert)

**Phase 1 — corren-sharia autonome (non destructif).** Créer le module : go.mod, package `chain`, déplacer la logique, casser `core`/`config`, son storage sqlite+postgres + migrations, ses handlers, ses tests (LedgerPort fake + store sqlite réel). `go test ./...` vert **dans corren-sharia**. corren intact.

**Phase 2 — bascule de corren (allègement).** `go.mod` require + `replace ../corren-sharia` ; adaptateur ledger ; rewire contract_controller/flows/scheduler/storage ; supprimer `corren/sharia`, `storage/{sqlite,postgres}/sharia.go`, migrations sharia ; `storage.Store` n'embarque plus `ShariaStore`. `go test ./...` vert **dans corren**, zéro régression.

## 6. Risques

- Type-boundary `chain ↔ core` : mitigé par des structs identiques + un adaptateur trivial, testé.
- Deux runners de migration sur un DB : tables disjointes + IF NOT EXISTS.
- `flows`/`scheduler`/controllers à recâbler : Phase 2 atomique, build vérifié.
- Intégrité transactionnelle conservée (Commit reste local au ledger via l'adaptateur, pas de réseau).

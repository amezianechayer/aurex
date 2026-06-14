# Design — FaRl Lens v1 (financial observability: analytics API + Horizon page)

**Date** : 2026-06-14
**Statut** : validé par le user (décisions figées ci-dessous)
**Base** : branche `feature/farl-lens-v1` (corren, empilée sur `feature/farl-guard-v1`)
+ une branche Horizon dérivée de `feature/contracts-ui`.
**Objectif** : l'observabilité financière de corren — agrégation **côté serveur**
(SQL) des flux, soldes, contrats et events, exposée par 4 endpoints en lecture,
et une page Lens dans Horizon. 100 % lecture : aucun impact sur `Commit`.

Jamais de commit sur `main`.

## 1. Architecture

Package `analytics/` (frère de `sharia/`, `guard/`) : types de résultats +
interface `AnalyticsStore`, embarquée dans `storage.Store` et implémentée par
`storage/sqlite/analytics.go` et `storage/postgres/analytics.go` (requêtes
`GROUP BY`/`SUM` — agrégation serveur, **jamais de fetch-all client**). Un
contrôleur `api/actions/lens_controller.go` expose 4 endpoints sous
`/:ledger/lens/...`. Le `analytics` package importe seulement `core` (pas de
cycle). Lens ne touche pas `ledger.Commit` ni aucune écriture.

## 2. Décisions figées

- **Découpage exécution** : un seul spec, mais exécuté en deux temps —
  **backend corren d'abord** (testable, autonome), **puis** l'UI Horizon qui le
  consomme. Évite la v1 qui traîne sur deux repos.
- **Contrat `/flows` (gelé, partagé table v1 + graphe phase 2)** : une ligne =
  `{ source, destination, asset, time_bucket, amount, count }`. Une seule forme.
  Le tableau v1 collapse les `time_bucket` par arête (agrégation client triviale,
  pas de fetch-all). Le graphe interactif (phase 2, hors v1) rejoue les buckets
  pour animer le flux — **zéro changement backend** entre les deux.
  `asset` est obligatoire (ledger multi-asset : on ne somme pas DZD.2 + EUR.2).
- **Graphe interactif = phase 2, hors v1.** Le coder en v1 couplerait Lens à la
  lib + design system de la branche `transaction-graph` et alourdirait la v1.
  v1 = tableau de flux soigné ; le graphe se rebranche plus tard sur le MÊME
  contrat `/flows`.
- **Lens est read-only.** Aucune mutation, aucun hook dans `Commit`.

## 3. Les 4 endpoints (sémantique figée)

Toutes les requêtes sont par-ledger, scopées à la base du ledger courant.

### 3.1 `GET /:ledger/lens/overview`
Vue d'accueil. Retourne :
```json
{ "transactions": 84, "accounts": 81,
  "volume_by_asset": [ {"asset":"DZD.2","total":12345600} ],
  "top_accounts": [ {"account":"@bank:treasury","asset":"DZD.2","volume":9000000} ] }
```
- `transactions`, `accounts` : réutilisent `CountTransactions`/`CountAccounts`.
- `volume_by_asset` : `SELECT asset, SUM(amount) FROM postings GROUP BY asset`.
- `top_accounts` (limit 10) : union de `(source AS account, asset, amount)` et
  `(destination AS account, asset, amount)`, `GROUP BY account, asset`,
  `ORDER BY SUM(amount) DESC LIMIT 10`.

### 3.2 `GET /:ledger/lens/flows?limit=N` (défaut 100, max 1000)
Le feed gelé. Retourne `[]{source, destination, asset, time_bucket, amount, count}` :
```sql
SELECT p.source, p.destination, p.asset,
       substr(t.timestamp,1,10) AS time_bucket,
       SUM(p.amount) AS amount, COUNT(*) AS count
FROM postings p JOIN transactions t ON p.txid = t.id
GROUP BY p.source, p.destination, p.asset, time_bucket
ORDER BY amount DESC
LIMIT N
```
(`substr(timestamp,1,10)` = `YYYY-MM-DD` — le timestamp est RFC3339 sur la table
`transactions`, d'où le join.) Le tableau v1 regroupe les lignes par
`(source,destination,asset)` en sommant les buckets ; le graphe phase 2 garde
les buckets pour l'animation temporelle.

### 3.3 `GET /:ledger/lens/rollup`
Le pont Lens↔Guard↔contrats. Retourne :
```json
{ "contracts_by_state": {"PROMISE":2,"SOLD":1,"SETTLED":3,"CANCELLED":0},
  "guard_events_by_action": {"deny":4,"monitor":7} }
```
- `SELECT state, COUNT(*) FROM sharia_contracts GROUP BY state`
- `SELECT action, COUNT(*) FROM guard_events GROUP BY action`
- Les tables existent toujours (créées aux migrations v002/v004). Si vides → maps vides.

### 3.4 `GET /:ledger/lens/timeseries?account=<acct>&asset=<asset>`
Flux entrée/sortie par jour d'un compte+asset (pour le graphe temporel) :
```json
[ {"time_bucket":"2026-07-01","in":500000,"out":0},
  {"time_bucket":"2026-07-31","in":0,"out":458333} ]
```
```sql
SELECT substr(t.timestamp,1,10) AS time_bucket,
       SUM(CASE WHEN p.destination = :acct THEN p.amount ELSE 0 END) AS in,
       SUM(CASE WHEN p.source      = :acct THEN p.amount ELSE 0 END) AS out
FROM postings p JOIN transactions t ON p.txid = t.id
WHERE (p.source = :acct OR p.destination = :acct) AND p.asset = :asset
GROUP BY time_bucket ORDER BY time_bucket ASC
```
`account` et `asset` requis (sinon `ERR_INVALID_PARAMS` 400).

## 4. AnalyticsStore (analytics/store.go)

```go
type AnalyticsStore interface {
    LensOverview() (Overview, error)
    LensFlows(limit int) ([]FlowEdge, error)
    LensRollup() (Rollup, error)
    LensTimeSeries(account, asset string) ([]TimeBucket, error)
}
```
Types (`analytics/types.go`) : `Overview{Transactions, Accounts int64;
VolumeByAsset []AssetVolume; TopAccounts []AccountVolume}`,
`FlowEdge{Source, Destination, Asset, TimeBucket string; Amount int64; Count int64}`,
`Rollup{ContractsByState, GuardEventsByAction map[string]int64}`,
`TimeBucket{TimeBucket string; In, Out int64}`. Embarquée dans `storage.Store`
(à côté de `sharia.ShariaStore`, `guard.GuardStore`).

## 5. Frontend Horizon (branche dérivée de feature/contracts-ui)

Stack inchangée (React 16 + styled-components). Nouvelle page **Lens**
(`/ledgers/:ledger/lens` ou `/lens`), lien Navbar :
- **Cartes overview** : transactions, comptes, volume par asset.
- **Rollup** : contrats par état + guard deny/monitor en compteurs colorés
  (réutilise `StateBadge`).
- **Série temporelle** : un graphe simple (SVG/CSS, zéro lib) in/out par jour
  pour un compte+asset sélectionné, depuis `/lens/timeseries`.
- **Tableau de flux soigné** (le livrable phare v1) depuis `/lens/flows` :
  top flux `source → destination` par asset, montant + count, **barres
  proportionnelles animées (grow-in), count-up des montants, transitions,
  hover**, tokens de design Horizon. Zéro dépendance externe. Ce tableau prouve
  et fige le contrat `/flows`.

Rôles distincts (pas de doublon) : tableau = workhorse lisibilité/audit ;
graphe (phase 2) = wow démo. Même endpoint, deux rendus.

## 6. Tests & critères d'acceptation

- **Backend (TDD)** : `analytics` + store. Pour chaque endpoint, un jeu de
  postings/contrats/guard_events connu et des valeurs **exactes** attendues
  (overview volumes, top accounts ordonnés, flow edges agrégées + buckets,
  rollup counts, timeseries in/out par jour). sqlite + **postgres gated**
  (`CORREN_TEST_PG_CONN`). httptest du contrôleur (4 endpoints, 200 + formes ;
  timeseries sans params → 400).
- **Données vides** : un ledger neuf → overview zéros, flows/timeseries vides,
  rollup maps vides (pas d'erreur).
- **Front** : QA navigateur (browse) — page Lens se charge, cartes/rollup/série/
  tableau s'affichent sur un ledger réel, captures avant/après.
- `go build ./... && go test ./...` verts ; `go vet` ; gofmt.

## 7. Hors périmètre v1 (→ phase 2 / v2)

Graphe interactif force-directed (réutilisation `transaction-graph`, animation
des arêtes, clustering) — pur frontend sur le contrat `/flows` identique.
Filtres temporels avancés, export CSV, alerting sur seuils, comparaison
multi-ledger, drill-down par transaction. Tous reportés.

## 8. Risques connus

- `/flows` time-bucketed peut renvoyer beaucoup de lignes (source×dest×asset×jour) :
  `limit` (défaut 100, max 1000) + `ORDER BY amount DESC` bornent. Un index
  existe déjà sur `postings(txid, source, destination)`. Optim (index dédié,
  pagination) = v2.
- `substr(timestamp,1,10)` suppose un timestamp RFC3339 (toujours posé par
  `Commit`). Vrai pour toutes les transactions du ledger.
- Périmètre « deux repos » : maîtrisé par le découpage backend-puis-UI (§2).

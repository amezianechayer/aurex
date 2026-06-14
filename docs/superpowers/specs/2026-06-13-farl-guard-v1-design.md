# Design — FaRl Guard v1 (règles déclaratives à `ledger.Commit`)

**Date** : 2026-06-13
**Statut** : validé par le user (décisions figées ci-dessous)
**Base** : branche `feature/farl-guard-v1`, empilée sur `feature/sharia-ijarah-v1`
(hérite : moteur multi-contrats, auth, postgres, fixes config/assets/legacy).
**Objectif** : une couche de règles déclaratives, configurable et hot-reloadable,
évaluée au **seul chokepoint** que toute mutation d'état traverse (`ledger.Commit`),
qui **bloque** (deny) ou **observe** (monitor) les transactions selon des règles
portant chacune leur motif et leur citation. Jamais de modification de corren-vm.

Jamais de commit sur `main`.

---

## 1. Pourquoi `ledger.Commit` (prouvé par inspection)

Les 4 émetteurs de postings convergent tous sur `ledger.Commit` :

| Émetteur | Chemin | corren-vm ? |
|---|---|---|
| Scripts FaRl (`ledger/executor.go`) | la VM résout `m.Postings` (ne persiste rien) puis `l.Commit` | résout, mais sort par Commit |
| Moteur de contrats (`sharia/engine.go`) | `e.ledger.Commit(...)` directement | **non** |
| API transaction directe (`api/actions/transaction_controller.go`) | `Commit` directement | **non** |
| Revert (`ledger/ledger.go`) | `Commit` directement | non |

Un pass niveau VM raterait le moteur de contrats ET l'API directe (2/4) et créerait
un 2ᵉ gate partiel. **Gate unique à `ledger.Commit`.** Le contrôle de balance
existant (`ledger.go`, boucle sur `rf`) est déjà la « règle Guard #0 » : Guard
généralise ce pattern (pass pré-persistance, `return ts, err` atomique avant
`SaveTransactions`).

## 2. Décisions figées

- **Forme des règles** : déclaratif (JSON), jamais en grammaire FaRl. Hot-reloadable.
- **Point d'enforcement** : `ledger.Commit`, juste avant `SaveTransactions`, même
  fenêtre atomique que le balance-check.
- **Modes** : `deny` (abort, zéro changement d'état) **et** `monitor`
  (journalise un guard_event, laisse passer — shadow/dry-run).
- **Frontière stricte Sharia ↔ Guard** : les invariants Sharia durs (qabd/SS-8,
  IJ-1..5, riba, charité SS-3…) restent dans `sharia/`, **toujours deny, jamais
  exprimables en monitor**. `guard/` = couche configurable (plafonds, blocklist,
  AML, juridiction). Une règle Sharia ne peut pas être dégradée en monitor.
- **Citation** : `standard_ref` free-form (AAOIFI / AML / interne `POLICY-X`).
  `reason` toujours requis. `standard_ref` **requis si `deny`**, optionnel si
  `monitor`. Un deny sans citation est interdit (refus inexpliqué = trou d'audit) —
  contrôlé à la validation de règle (impossible d'enregistrer un deny sans réf).
- **Critère de coupe v1 (vaut pour toute règle future)** : une règle v1 est
  évaluable depuis **`proposedTxs + netFlows` seuls**. Les balances ne sont
  garanties chargées que pour les comptes **net-sortants** (le balance-check ne
  charge `AggregateBalances` que pour eux, à la demande). Toute règle nécessitant
  un accès hors fenêtre (historique, scan, balance d'un net-entrant) → v2.
- **Durabilité (figé)** :
  - `deny` → écrit son guard_event par un **Exec store autonome** juste avant
    `return err`. Comme `SaveTransactions` n'est jamais appelé sur un deny (pas de
    transaction DB englobante dans ce codebase), l'event persiste forcément. Si cet
    Exec échoue : on **return quand même l'erreur de deny** (la tx ne passe jamais)
    en loggant l'échec secondaire. La preuve survit au rollback.
  - `monitor` → écrit juste **après** un `SaveTransactions` réussi (best-effort,
    écriture séparée). Même modèle de durabilité que la piste d'audit Sharia
    existante (`engine.go` : Commit puis appendAudit séparé). Atomicité stricte
    monitor↔commit = durcissement v2.
- **guard_events** : append simple (`id` + `created_at`) en v1 ; chaîne hashée
  tamper-evident en v2 (la preuve survit déjà au rollback ; le chaînage est un
  durcissement séparé).
- **Concurrence** : set de règles en mémoire, lu en **snapshot sous RWMutex** à
  chaque `Evaluate`, pour que le hot-reload (via l'API admin sur une autre
  goroutine) ne race pas une évaluation en cours. (Note : `Commit` tient déjà le
  mutex global du ledger, donc deux évaluations sur le même ledger ne sont jamais
  concurrentes ; le RWMutex protège évaluation vs reload.)

## 3. Types de règles v1

Toutes évaluables depuis `proposedTxs + netFlows` seuls :

1. **Plafond de montant (`amount_cap`)** — refuse, pour un asset donné, si :
   - `basis: "posting"` → un posting individuel dont la **source** matche `scope`
     a un `amount` > `max` ; ou
   - `basis: "net_outflow"` → le net sortant d'un compte matchant `scope` dépasse
     `max`. Le net sortant se lit directement dans `netFlows` (= la map `rf` de
     `Commit`, convention : `rf[source] += amount`, `rf[dest] -= amount`, donc
     `net_outflow(acct, asset) = max(0, rf[acct][asset])`).
   Params : `{scope, asset, max, basis: "posting"|"net_outflow"}`.
2. **Blocklist / allowlist (`account_list`)** — refuse (blocklist) ou n'autorise
   que (allowlist) des postings dont la source/destination matche un motif.
   Params : `{mode: "block"|"allow", side: "source"|"destination"|"either", patterns: [...]}`.
3. **Restriction d'asset (`asset_restrict`)** — un compte matchant le scope ne peut
   transiger que (ou jamais) certains assets.
   Params : `{scope, mode: "only"|"never", assets: [...]}`.

Reporté v2 : **vélocité / rate-limit** (fenêtre glissante → scan d'historique,
hors fenêtre Commit → son propre design perf).

Le **scope** et les **patterns** sont des motifs de compte façon glob simple
(`@client:*`, `@contracts:*`, `@bank:treasury`) — matching par préfixe avec `*`
terminal, pas de regex (YAGNI, et lexer-safe).

## 4. Modèle de données (migrations `v004`, deux flavors)

Par-ledger, style existant (`--statement`, `CREATE TABLE IF NOT EXISTS`,
`integer`→`bigint` pour Postgres).

```sql
--statement
CREATE TABLE IF NOT EXISTS guard_rules (
  "id"           varchar,      -- "rule-..." (fourni ou généré)
  "kind"         varchar,      -- amount_cap | account_list | asset_restrict
  "params"       varchar,      -- JSON spécifique au kind
  "action"       varchar,      -- deny | monitor
  "reason"       varchar,      -- requis
  "standard_ref" varchar,      -- requis si action=deny
  "enabled"      integer,      -- 0/1
  "created_at"   varchar,
  "updated_at"   varchar,
  UNIQUE("id")
);
--statement
CREATE TABLE IF NOT EXISTS guard_events (
  "id"           integer,      -- séquence par-ledger (count+1)
  "rule_id"      varchar,
  "action"       varchar,      -- deny | monitor (le mode de la règle qui a tiré)
  "reason"       varchar,
  "standard_ref" varchar,
  "tx_reference" varchar,      -- Reference de la tx concernée si présente
  "payload"      varchar,      -- JSON : postings/comptes/montants en cause
  "created_at"   varchar
);
--statement
CREATE INDEX IF NOT EXISTS 'guard_events_created' ON "guard_events" ("created_at");
```

`GuardStore` (interface dans `guard/store.go`, branchée sur `storage.Store` comme
`ShariaStore`) : `SaveRule`, `UpdateRule`, `DeleteRule`, `GetRule`, `ListRules`,
`AppendGuardEvent`, `ListGuardEvents(limit, offset)`. Implémentée dans
`storage/sqlite/guard.go` et `storage/postgres/guard.go`.

## 5. Le moteur (`guard/engine.go`)

```go
type Engine struct {
    store GuardStore
    mu    sync.RWMutex
    rules []Rule        // snapshot en mémoire, hot-reloadable
}

// Reload recharge le set depuis le store (appelé au démarrage et après tout
// changement via l'API admin).
func (e *Engine) Reload() error

// Evaluate est le hook appelé par ledger.Commit. proposedTxs = les tx en cours
// de commit ; netFlows = la map rf déjà calculée par Commit (compte -> asset ->
// net). Retourne une *guard.Error (deny) ou nil. Écrit les guard_events :
//   - deny  : AppendGuardEvent immédiat (Exec autonome) PUIS retour de l'erreur
//   - monitor: collecte les events ; ils sont écrits par le ledger APRÈS un
//     SaveTransactions réussi (Evaluate retourne la liste via un second canal).
func (e *Engine) Evaluate(led LedgerView, proposedTxs []core.Transaction,
    netFlows map[string]map[string]int64) (monitorEvents []GuardEvent, err error)
```

`LedgerView` = surface minimale (lecture de compte si une règle future en a
besoin) pour éviter le cycle d'import `guard → ledger → … → guard`, même pattern
que `sharia.LedgerPort`. En v1 les 3 règles n'ont pas besoin de `LedgerView`,
mais on le passe pour la stabilité de signature.

Algorithme `Evaluate` : prendre un snapshot `rules` sous `RLock` ; pour chaque
règle `enabled`, matcher contre `proposedTxs`/`netFlows` ; à la **première** règle
`deny` qui matche → écrire le guard_event de deny (Exec autonome via le store) →
retourner la `*guard.Error` (HTTP 422, format type `sharia.Error` :
`{error, message, standard_ref, rule_id}`). Les règles `monitor` qui matchent
sont accumulées dans `monitorEvents` (pas d'arrêt). S'il n'y a aucun deny,
retourner `(monitorEvents, nil)`.

## 6. Intégration dans `ledger.Commit` (chirurgicale)

Dans `ledger/ledger.go`, après la boucle de balance-check et **avant**
`l.store.SaveTransactions(ts)` :

```go
monitorEvents, gerr := l.guard.Evaluate(l, ts, rf)
if gerr != nil {
    return ts, gerr      // deny : SaveTransactions jamais appelé, _last non avancé
}
err := l.store.SaveTransactions(ts)
if err == nil && len(monitorEvents) > 0 {
    l.guard.WriteMonitorEvents(monitorEvents)  // best-effort, post-commit
}
l._last = &ts[len(ts)-1]
return ts, err
```

Le `guard.Engine` est injecté dans le `Ledger` à la construction
(`NewLedger`), `Reload()` au démarrage. **Défaut : aucune règle → `Evaluate`
retourne `(nil, nil)` immédiatement → comportement actuel strictement inchangé.**

## 7. API (`api/actions/guard_controller.go`)

Routes par-ledger, dans le groupe `/:ledger`, rôle `admin` (réutilise l'auth
existante : mutations de règles = admin ; `GET /events` = readonly+) :

```
POST   /:ledger/guard/rules            créer une règle (valide : deny ⇒ standard_ref requis)
GET    /:ledger/guard/rules            lister
GET    /:ledger/guard/rules/:id        détail
PATCH  /:ledger/guard/rules/:id        activer/désactiver / modifier
DELETE /:ledger/guard/rules/:id        supprimer
GET    /:ledger/guard/events           journal des déclenchements (deny+monitor)
```

Toute mutation de règle déclenche `guard.Engine.Reload()` du ledger concerné
(hot-reload). Création/modif d'une règle `deny` sans `standard_ref` → 400.

## 8. Erreurs typées (`guard/errors.go`)

| Code | HTTP | Cas |
|---|---|---|
| `ERR_GUARD_DENIED` | 422 | une règle deny a bloqué la tx (réponse : reason + standard_ref + rule_id) |
| `ERR_INVALID_RULE` | 400 | règle malformée (ex. deny sans standard_ref, kind inconnu) |
| `ERR_NOT_FOUND` | 404 | règle inconnue |

Format de réponse deny (identique en esprit à `sharia.Error`) :
```json
{ "error": "ERR_GUARD_DENIED", "message": "amount exceeds cap",
  "standard_ref": "POLICY-LIMIT-1", "rule_id": "rule-abc" }
```

## 9. Tests & critères d'acceptation

- **Unitaires** (`guard/rules_test.go`) : chaque kind — match/no-match sur des
  `proposedTxs`/`netFlows` synthétiques ; validation de règle (deny sans
  standard_ref → ERR_INVALID_RULE) ; matching de scope glob (`@client:*`).
- **Store** (`storage/sqlite/guard_test.go`) : round-trip rule + event ; migration v004.
- **Intégration** (réutilise le pattern `ledger`/`with`) :
  - **tx conforme** → commit normal (aucune règle, ou règle qui ne matche pas).
  - **deny** → `Commit` retourne `ERR_GUARD_DENIED` + reason + standard_ref ;
    **ledger inchangé** (la tx n'est pas dans le store, balances inchangées) ;
    **guard_event de deny persisté MALGRÉ l'absence de SaveTransactions**.
  - **monitor** → la tx passe ; guard_event monitor écrit après le commit.
  - **scoping `@contracts:*`** : une règle deny scope `@contracts:*` bloque un
    posting généré par le **moteur de contrats** (preuve que Guard couvre les
    mutations du moteur, pas seulement l'API/scripts).
  - **hot-reload concurrent** : `Reload()` pendant des `Evaluate` en boucle (race
    detector `-race`) → pas de data race.
  - **zéro règle = no-op** : un ledger sans règle se comporte exactement comme
    avant (un test compare un flux Murabaha complet avec/sans guard engine).
- `go build ./...` et `go test ./...` verts. `go test -race ./guard/ ./ledger/`.

## 10. Frontière de responsabilité (rappel)

`sharia/` = invariants religieux **non négociables**, toujours deny, codés en dur,
testés A–G / IJ-1..5. `guard/` = politique **configurable** de l'opérateur
(plafonds, listes, assets, AML), deny **ou** monitor. Les deux gates coexistent
dans `Commit` : balance-check (natif) → guard.Evaluate (configurable). Le moteur
de contrats produit des postings qui passent par les deux, ce qui rend Guard
capable d'encadrer aussi les flux contrats — le **mode monitor + le scoping**
sont les garde-fous pour introduire une règle sans casser un flux contrat
légitime.

## Hors périmètre v1

Vélocité / rate-limit (fenêtre glissante), chaîne hashée des guard_events,
atomicité stricte monitor↔commit, UI Horizon des règles, règles inter-tx
(corrélation sur plusieurs commits), import de listes AML externes. Tous → v2.

## Risques connus

- Une règle deny mal scopée peut bloquer un flux du moteur de contrats (créance
  Murabaha 11M sous un plafond). Mitigation : opt-in + scoping + **monitor d'abord**
  (déployer en monitor, observer ce qui aurait été bloqué, puis activer en deny).
- `Evaluate` est sur le chemin chaud de chaque commit : garder O(règles × postings),
  snapshot en mémoire, zéro I/O sauf l'écriture d'un guard_event quand une règle
  tire. Acceptable v1 ; un index/compilation des règles est une optim v2.

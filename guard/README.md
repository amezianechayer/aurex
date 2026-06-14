# guard — FaRl Guard: declarative rules at the ledger chokepoint

Guard is a configurable, hot-reloadable rule layer evaluated at the **single
chokepoint every state mutation crosses**: `ledger.Commit`. A rule either
**denies** a transaction (abort, zero state change) or **monitors** it (log,
allow). Each rule carries its own reason and citation.

## Why `ledger.Commit` (and not corren-vm)

Every posting funnels through `ledger.Commit`: FaRl scripts (the VM resolves
postings, then commits), the contracts engine (`sharia.Engine` commits
directly), the direct `/transactions` API, and reverts. A pass inside corren-vm
would miss the contracts engine and the direct API (2 of the 4 emitters) and
create a second partial gate. Guard runs at `Commit`, right after the existing
balance check (which is itself "Guard rule #0") and **before**
`SaveTransactions`. On deny, `Commit` returns the error → `SaveTransactions`
never runs → zero state change.

## Deny vs monitor

| Action | Effect | Event durability |
|---|---|---|
| `deny` | abort the commit; the transaction never persists | the deny `guard_event` is written by its own store call **before** the abort, so the proof survives even though the transaction does not |
| `monitor` | record a `guard_event`, allow the transaction | written best-effort after a successful commit (same durability model as the sharia audit trail) |

Monitor is the safe-rollout / shadow mode: deploy a rule in monitor, watch what
it *would* block (including contract-engine flows), then switch it to deny.

## Sharia ↔ Guard boundary

Hard religious invariants (qabd/SS-8, IJ-1..5, riba, charity/SS-3) live in
`sharia/`, are **always deny**, and can never be expressed as monitor rules.
`guard/` is the operator's configurable policy layer (limits, lists, assets,
AML, jurisdiction). A Sharia invariant cannot be downgraded to monitor — they
are different packages with different stores.

## Rule kinds (v1)

All evaluable from `proposedTxs + netFlows` alone (no extra reads). Scopes and
patterns are simple prefix globs: `@client:*`, `@contracts:*`, `@bank:treasury`,
or `*`.

```json
// amount_cap — block a posting (or net outflow) above a cap, per asset
{ "kind": "amount_cap", "action": "deny", "reason": "tx limit",
  "standard_ref": "POLICY-LIMIT-1",
  "params": { "scope": "@client:*", "asset": "DZD.2", "max": 100000000,
              "basis": "posting" } }   // basis: "posting" | "net_outflow"

// account_list — block (or allow-only) source/destination patterns
{ "kind": "account_list", "action": "deny", "reason": "sanctions",
  "standard_ref": "AML-SDN",
  "params": { "mode": "block", "side": "destination",
              "patterns": ["@sanctioned:*"] } }  // mode: block|allow; side: source|destination|either

// asset_restrict — a scoped account may only / may never transact some assets
{ "kind": "asset_restrict", "action": "monitor", "reason": "watch gold",
  "params": { "scope": "@client:*", "mode": "never", "assets": ["GOLD"] } }
```

Velocity / rate-limit needs a rolling-window history scan (outside the Commit
window) → v2, with its own performance design.

## Validation

`reason` is always required. **`standard_ref` is required for a `deny` rule**
(an unexplained refusal is an audit hole; even an internal `POLICY-X` ref is
enough). Checked at rule-save time — a deny rule without a citation is rejected
with `ERR_INVALID_RULE` (400).

## Concurrency

The rule set is held in memory and read as a snapshot under `RWMutex` on every
`Evaluate`, so a hot-reload (triggered by an admin mutation on another
goroutine) never races an in-flight evaluation.

## API (per ledger, admin)

```
POST   /:ledger/guard/rules            create a rule (deny ⇒ standard_ref required)
GET    /:ledger/guard/rules            list rules
GET    /:ledger/guard/rules/:id        get a rule
PATCH  /:ledger/guard/rules/:id        update / enable / disable
DELETE /:ledger/guard/rules/:id        delete
GET    /:ledger/guard/events           the firing log (deny + monitor)
```

Every rule mutation triggers `engine.Reload()` for that ledger (hot-reload).

Deny error shape (HTTP 422 at the engine/ledger level):
```json
{ "error": "ERR_GUARD_DENIED", "message": "tx limit",
  "standard_ref": "POLICY-LIMIT-1", "rule_id": "rule-..." }
```
(The generic `/transactions` controller maps any commit error to 500; the typed
`ERR_GUARD_DENIED` is what `ledger.Commit` returns and what the contracts engine
propagates.)

## Default: no rules = no-op

A ledger with no rules makes `Evaluate` return immediately — behavior is
strictly identical to before Guard existed.

## Tables (migration v004)

`guard_rules` (declarative, hot-reloaded) and `guard_events` (every firing,
append-only — the observability feed). v1 events are a plain append
(`id` + `created_at`); a tamper-evident hash chain is a v2 hardening (the deny
proof already survives the rollback regardless).

# analytics — FaRl Lens: financial observability

Read-only, server-side aggregation over the corren ledger. Four endpoints under
`/:ledger/lens/...`. All aggregation happens in SQL (`GROUP BY`/`SUM`) — never a
client-side fetch-all. Lens never writes and has no hook in `ledger.Commit`.

The `analytics` package holds the result types + the `AnalyticsStore` interface,
embedded into `storage.Store` (like `sharia.ShariaStore` and `guard.GuardStore`)
and implemented by `storage/sqlite` and `storage/postgres` with identical
exact-value tests on both flavors.

## Endpoints

### `GET /:ledger/lens/overview`
Counts + volume per asset + top-10 accounts by volume.
```json
{ "transactions": 84, "accounts": 81,
  "volume_by_asset": [ {"asset":"DZD.2","total":1400} ],
  "top_accounts": [ {"account":"@client:a","asset":"DZD.2","volume":1400} ] }
```

### `GET /:ledger/lens/flows?limit=N`  (default 100, max 1000)
The **frozen flow contract**, shared by the v1 ranked table and the phase-2
interactive graph — one shape, two renders. Each row is a time-bucketed edge:
```json
[ {"source":"@world","destination":"@client:a","asset":"DZD.2",
   "time_bucket":"2026-07-01","amount":1000,"count":1} ]
```
`asset` is mandatory (multi-asset ledgers never sum across assets). The table
collapses `time_bucket` per `(source,destination,asset)`; the graph replays
buckets to animate flow over time — **zero backend change between the two**.

### `GET /:ledger/lens/rollup`
The Lens↔Guard↔contracts bridge.
```json
{ "contracts_by_state": {"PROMISE":1,"SOLD":0,"SETTLED":0},
  "guard_events_by_action": {"deny":1,"monitor":0} }
```

### `GET /:ledger/lens/timeseries?account=<acct>&asset=<asset>`
Daily in/out for one account+asset (both params required, else 400).
```json
[ {"time_bucket":"2026-07-01","in":1000,"out":0},
  {"time_bucket":"2026-07-02","in":0,"out":400} ]
```

## Empty ledgers
A fresh ledger returns zeros / empty arrays / non-nil empty maps — never an error.

## Out of scope (phase 2 / v2)
Interactive force-directed flow graph (reuses the horizon `transaction-graph`
work, on the same `/flows` contract — pure frontend, no backend change),
advanced time filters, CSV export, threshold alerting, multi-ledger comparison.

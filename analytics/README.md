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

## Flow graph (phase 2) — built, pure frontend on `/flows`
A second render of the frozen `/flows` contract (the first being the v1 ranked
table), shipped on the Horizon branch `feature/lens-ui` with **zero backend
change**. It is a d3-v7 force-directed graph of money flow:

- **Pure transforms** (`horizon/src/lib/buildGraph.js`): `buildGraph(flows, asset)`
  → `{nodes, links, buckets}`; `kindOf(account)` maps the address prefix to a
  semantic kind (`bank|client|supplier|contracts|world|other`); node `volume` =
  sum of incident amounts; links collapse `(source, destination)`.
- **Single asset at a time** via a selector (default = largest-volume asset);
  thicknesses are never compared across assets.
- **Automatic clustering by kind**: a per-kind centroid force groups nodes; any
  kind can collapse into one labelled meta-node ("Clients (42)") with aggregated
  edges. Default-collapsed when a kind has > 12 members. Collapse/expand is driven
  by reliable HTML legend buttons (not clicks on moving SVG nodes).
- **Cumulative time-scrubber** (`filterToBucket`): position T shows all flows up
  to T; play/pause replays the timeline (700 ms/bucket) and restarts from the
  first bucket when pressed at the end.
- **Robustness**: the d3 simulation is stopped and the SVG cleared on every
  asset/limit/cluster/bucket change (no post-unmount ticking); no `setState` in
  the d3 tick; empty ledgers render "Aucun flux à afficher." with no graph mounted.

## Out of scope (v2)
Advanced time filters, windowed (non-cumulative) view, CSV/PNG export, threshold
alerting, multi-ledger comparison, connectivity-based community detection.

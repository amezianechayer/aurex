# sharia — Islamic finance contract engine

AAOIFI-aligned Islamic finance contracts on top of the corren programmable
ledger. API-first: every state change is one atomic ledger transaction plus
one hash-chained audit event. All amounts are `int64` minor units — no floats,
anywhere.

Two contracts ship today: **Murabaha** (cost-plus sale) and **Ijarah**
(operating lease). The engine is multi-contract: see "Adding a contract" below.

## State machine

```
PROMISE ──acquire──▶ ACQUIRED ──sell──▶ SOLD ──pay_installment (×N)──▶ SETTLED
   │                    │                 │
   └──cancel──▶ CANCELLED◀───cancel───────┤  (cancel from SOLD: forbidden, C-2)
                                          └──early_settle──▶ SETTLED
```

Any transition outside this graph is rejected with `ERR_INVALID_TRANSITION`
**and journalized** — a refusal is evidence, not a silent error.

## Ledger accounts (per contract `<id>`)

| Account | Role | May go negative |
|---|---|---|
| `@contracts:<id>:inventory` | bank's possession of the asset (qabd) | no |
| `@contracts:<id>:receivable` | client's outstanding debt | no |
| `@contracts:<id>:deferred` | unearned profit (FAS 28) | no |
| `@contracts:<id>:counterpart` | technical balancing counterpart | **yes** (only exemption) |
| `@bank:treasury` | real cash | no |
| `@bank:income:murabaha` | recognized Murabaha income | no |
| `@charity:pool` | sole lawful destination of late penalties | no |

At any moment: `balance(receivable) == Σ unpaid installment amounts` and
`balance(deferred) == Σ unrecognized profit parts`.

## Invariants → standards → tests

| Invariant | Rule | Standard | Test |
|---|---|---|---|
| I-1 | No sale without possession (qabd): `sell` requires the asset in the contract inventory | AAOIFI-SS-8 | `TestScenarioBSellWithoutPossession` |
| I-2 | `cost` and `markup` immutable after creation; markup is a fixed amount, never a rate | AAOIFI-SS-8 | no mutating endpoint exists; `TestValidateParamsRejects` |
| I-3 | Late penalties may only flow to `@charity:*`, never to bank income | AAOIFI-SS-3 | `TestScenarioCPenaltyToIncomeRejected`, `TestPenaltyPostingsRejectsNonCharityDestination` |
| I-4 | Profit recognized pro rata as installments are paid, never upfront | AAOIFI-FAS-28 | `TestScenarioANominalCycle`, `TestScenarioDEarlySettle` |
| I-5 | Strict FSM sequencing; every refusal journalized | — | `TestScenarioFInvalidTransitions`, `TestFSMAllowedTransitions` |
| I-6 | Integer arithmetic to the cent; remainder on the last installment | — | `TestSplitEven`, `TestBuildScheduleReference` |
| I-7 | One atomic ledger transaction per transition, or nothing | — | `TestScenarioEIdempotence` (unique `Reference`) |

Standard references are document-level (`AAOIFI-SS-8`, `AAOIFI-SS-3`,
`AAOIFI-FAS-28`); exact clause numbers must be confirmed by a Sharia
advisor before production.

---

## Ijarah (operating lease)

The bank buys an asset, keeps it on its own books (continuous ownership is
what makes the rent licit), leases the *use* to the client, recognizes rent
period by period, depreciates the asset straight-line, and takes the asset
back at the end. No ownership transfer (lease-to-own / IMB is v2).

### State machine

```
PROMISE ──acquire──▶ ACQUIRED ──lease──▶ LEASED ──pay_rent (×N)──▶ COMPLETED
   │                    │                   │
   └──cancel──▶ CANCELLED◀──cancel          └──(asset returned to the lessor)
```

`lease` ≠ a sale: only possession (usufruct) moves to the client; ownership
and the book value stay with the bank. `cancel` from LEASED is forbidden.

### Ledger accounts (per contract `<id>`)

| Account | Role | May go negative |
|---|---|---|
| `@contracts:<id>:asset` | leased asset: the physical unit **and** its book value, depreciating. Lessor-owned throughout. | no |
| `@client:<x>:in_use` | the client's possession of the unit during the lease | no |
| `@bank:income:ijarah` | recognized rental income | no |
| `@bank:expense:depreciation` | accumulated depreciation | no |
| `@bank:inventory:returned` | where the unit lands at lease end | no |

Ijarah does **not** use `:counterpart` — it has no upfront receivable, so the
book value and recognized rent are sourced from `@world` (the ledger's
convention for value entering an account). At completion: treasury holds the
margin (`Σ rent − cost`), `:asset` book value is 0, income = `Σ rent`,
depreciation = `cost`, counterpart = 0.

### Invariants → standards → tests

| Invariant | Rule | Standard | Test |
|---|---|---|---|
| IJ-1 | No lease without ownership: `lease` requires the asset in `:asset` | AAOIFI-SS-9 | `TestIjarahNominalCycle` (lease-before-acquire → 422) |
| IJ-2 | Lessor stays the owner: the unit never transfers in ownership; book value stays on the books | AAOIFI-SS-9 | `TestIjarahNominalCycle` |
| IJ-3 | Rent recognized period by period, never a debt born upfront | FAS-32 (simplified) | `TestIjarahNominalCycle` (no receivable at lease) |
| IJ-4 | Straight-line depreciation, labelled "FAS 32 simplified (v1)", basis advisor-to-validate | FAS-32 (simplified) | `TestBuildIjarahScheduleReference` |
| IJ-5 | Late penalties → charity only | AAOIFI-SS-3 | `TestIjarahPenaltyToIncomeRejected` |

The accounting is labelled **"FAS 32 — simplified (v1)"**, never "FAS 32
compliant"; the depreciation basis must be validated by a Sharia/accounting
advisor before production.

**v1 limitation:** the asset return shares the last `pay_rent` transition, so
it has no standalone audit event (not separately queryable). Splitting it out
is a v2 refinement.

## Adding a contract (multi-contract engine)

The engine routes by `Contract.Type` through `registry map[string]ContractKind`.
A contract type is one `ContractKind` implementation (`sharia/murabaha.go`,
`sharia/ijarah.go`) that supplies: param decode/validate, the schedule, the
FSM, a pre-FSM `ShariaGate`, `Preconditions`, and a `BuildPlan` returning the
postings + state + audit. Register it with `init() { register(yourKind{}) }`.
The engine (`engine.go`) holds zero contract-specific knowledge — adding a
contract never touches it.

## Audit chain

Each contract has its own chain: `prev_hash(0) = sha256(contract_id)`,
`hash(s) = sha256(prev_hash || canonical_payload)`. Events link the ledger
`tx_id` (itself hash-chained by `Commit`) — double anchoring.
`GET /:ledger/contracts/:id/audit?verify=true` re-walks the whole chain.

## API

```
POST /:ledger/contracts                          create (PROMISE, no posting)
GET  /:ledger/contracts                          list
GET  /:ledger/contracts/:id                      contract + schedule
POST /:ledger/contracts/:id/transitions/:name    acquire | sell | pay_installment |
                                                 early_settle | late_penalty | cancel
GET  /:ledger/contracts/:id/schedule             installment schedule
GET  /:ledger/contracts/:id/audit[?verify=true]  audit trail (+ chain_valid)
```

Errors: `ERR_INVALID_PARAMS` 400, `ERR_NOT_FOUND` 404,
`ERR_INVALID_TRANSITION`/`ERR_DUPLICATE` 409,
`ERR_PRECONDITION`/`ERR_SHARIA_VIOLATION` 422 (with `standard_ref`).

Authentication: when `auth.enabled=true`, every call needs
`Authorization: Bearer <token>` (api key or session). Roles: `readonly`
(GET only), `operator` (transitions), `admin`. Bootstrap with
`corren auth init`. See the `auth/` package.

See `demo/murabaha_demo.sh` for the full curl walkthrough.

## Testing against Postgres

The sqlite suite runs everywhere; the postgres paths are covered by
env-gated integration tests:

```bash
docker compose up -d postgres   # exposed on host port 5433
CORREN_TEST_PG_CONN="postgresql://ledger:ledger@localhost:5433/ledger" \
  go test ./storage/postgres/ ./auth/ ./sharia/ -count=1
```

Without the variable these tests skip and `go test ./...` needs no database.

## Scheduler

A background goroutine (`sharia.scheduler.interval`, default 1h) marks
`pending` installments past `sharia.grace_days` (default 7) as `overdue` on
SOLD contracts and journalizes them. It never posts to the ledger;
penalties remain an explicit API decision in v1.

# sharia — Islamic finance contract engine (v1: Murabaha)

AAOIFI-aligned Murabaha contracts on top of the corren programmable ledger.
API-first: every state change is one atomic ledger transaction plus one
hash-chained audit event. All amounts are `int64` minor units — no floats,
anywhere.

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

See `demo/murabaha_demo.sh` for the full curl walkthrough.

## Scheduler

A background goroutine (`sharia.scheduler.interval`, default 1h) marks
`pending` installments past `sharia.grace_days` (default 7) as `overdue` on
SOLD contracts and journalizes them. It never posts to the ledger;
penalties remain an explicit API decision in v1.

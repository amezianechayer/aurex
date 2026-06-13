# Ijarah v1 + Multi-Contract Engine — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generalize the sharia engine behind a `ContractKind` interface (behavior-preserving for Murabaha), then add operating Ijarah as a second contract type.

**Architecture:** `Contract.Params` becomes `json.RawMessage`; a `registry map[string]ContractKind` holds one definition per contract type. `Engine.Transition`/`Create` route by `Contract.Type` to the kind, which decodes its own params and supplies FSM, a pre-FSM Sharia gate, preconditions and a posting plan. Murabaha logic is extracted into `murabahaKind` with zero behavior change; the A–G suite is the safety net. Ijarah is then a new `ijarahKind`.

**Tech Stack:** Go 1.16, gin, go-sqlbuilder, fx. No new dependency.

**Spec:** `docs/superpowers/specs/2026-06-13-sharia-ijarah-v1-design.md`
**Branch:** `feature/sharia-ijarah-v1` (already created). NEVER commit to main.

**Golden rule for Part 1:** the refactor is behavior-preserving. Existing tests
(`./sharia`, `./api/...`, `./storage/...`, `./scheduler`, A–G scenarios) must
stay green and UNCHANGED. If one breaks, fix the refactor, not the test.

---

## File map

| File | Change | Responsibility |
|---|---|---|
| `sharia/kind.go` | create | `ContractKind` interface, `Params`, `Event`, `TransitionPlan`, `InstallmentMark`, registry |
| `sharia/murabaha.go` | modify | add `murabahaKind` implementing the interface (extract from engine) |
| `sharia/types.go` | modify | `Contract.Params` → `json.RawMessage`; keep `MurabahaParams`; add `IjarahParams`, ijarah states/transitions/consts |
| `sharia/engine.go` | modify | generic `Transition`/`Create` routing via registry; keep audit/lock/commit plumbing |
| `sharia/schedule.go` | modify | add `BuildIjarahSchedule`; keep `SplitEven`, `BuildSchedule` |
| `sharia/ijarah.go` | create | `ijarahKind` implementation |
| `sharia/ijarah_test.go` | create | Ijarah unit + scenario tests |
| `sharia/types.go` (Installment) | modify | reuse `Installment`; `ProfitPart` doubles as `depreciation_part` for ijarah rows (documented), or add a generic field — see Task 7 |
| `sharia/README.md` | modify | Ijarah section + multi-contract note |

---

## PART 1 — Behavior-preserving multi-contract refactor

### Task 1: Introduce the ContractKind interface and registry (no wiring yet)

**Files:** Create `sharia/kind.go`

- [ ] **Step 1: Write `sharia/kind.go`** (new types only, compiles, nothing uses them yet)

```go
package sharia

import (
	"encoding/json"

	"github.com/amezianechayer/corren/core"
)

// Params is an opaque marker; each ContractKind knows its concrete type.
type Params interface{}

// Event is a transition request: a name plus the HTTP-provided input.
type Event struct {
	Name  string
	Input TransitionInput
}

// InstallmentMark is a declarative post-commit hook: mark one installment.
type InstallmentMark struct {
	Seq    int
	Status string
}

// TransitionPlan is what a kind asks the engine to commit, plus the
// declarative post-commit bookkeeping the engine performs generically.
type TransitionPlan struct {
	Postings    []core.Posting
	Reference   string
	NewState    string
	StandardRef string
	Event       string // audit event kind; defaults to EventTransition
	Payload     string
	Marks       []InstallmentMark // installments to mark paid/settled/etc.
	ExtraAudit  []AuditEvent      // additional audit events (e.g. settled)
}

// ContractKind defines one contract type. Adding a contract = adding a kind
// and registering it; the engine never changes.
type ContractKind interface {
	Type() string
	DecodeParams(raw json.RawMessage) (Params, error)
	ValidateParams(p Params) error
	BuildSchedule(p Params) ([]Installment, error)
	// ShariaGate runs BEFORE the FSM check so a possession/ownership
	// violation takes priority over a sequencing error. nil = no gate.
	ShariaGate(led LedgerPort, c Contract, p Params, ev Event) error
	AllowedTransitions(from string) []string
	Preconditions(led LedgerPort, c Contract, p Params, sched []Installment, ev Event) error
	BuildPlan(led LedgerPort, c Contract, p Params, sched []Installment, ev Event) (TransitionPlan, error)
}

var registry = map[string]ContractKind{}

func register(k ContractKind) { registry[k.Type()] = k }

func kindFor(contractType string) (ContractKind, bool) {
	k, ok := registry[contractType]
	return k, ok
}
```

- [ ] **Step 2: Verify build + all tests green**

Run: `go build ./... && go test ./... -count=1`
Expected: builds; every package `ok` (nothing wired yet, pure addition).

- [ ] **Step 3: Commit**

```bash
git add sharia/kind.go
git commit -m "feat(sharia): add ContractKind interface and registry (unwired)"
```

---

### Task 2: Change Contract.Params to json.RawMessage

**Files:** Modify `sharia/types.go`, `sharia/engine.go`, `sharia/store.go` consumers, `storage/*/sharia.go`

The store already serializes `Params` as JSON, so the DB is unchanged. Only Go
types change. `MurabahaParams` stays; `Contract` stops embedding it directly.

- [ ] **Step 1: Edit `sharia/types.go`** — change the field type

```go
type Contract struct {
	ID              string          `json:"id"`
	Type            string          `json:"type"`
	State           string          `json:"state"`
	Params          json.RawMessage `json:"params"`
	TemplateVersion string          `json:"template_version"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}
```

Add `import "encoding/json"` to types.go if not present.

- [ ] **Step 2: Build to find every break**

Run: `go build ./... 2>&1 | head -40`
Expected: compile errors everywhere `c.Params.X` is accessed (engine.go) and
where `Contract{Params: someMurabahaParams}` is constructed (engine.go,
storage tests). This is the worklist for Steps 3–4.

- [ ] **Step 3: Fix the storage layer** — `storage/sqlite/sharia.go` and
`storage/postgres/sharia.go` currently `json.Marshal(c.Params)` on save and
`json.Unmarshal(..., &c.Params)` on load. With `Params` now `json.RawMessage`:
- on save: `params := []byte(c.Params)` (already JSON; if empty, use `[]byte("{}")`).
- on load: `c.Params = json.RawMessage(paramsString)`.

Apply in both flavors (`SaveContract`, `GetContract`, `ListContracts`).

- [ ] **Step 4: Fix storage tests** — `storage/sqlite/sharia_test.go` and
`storage/postgres/sharia_test.go` build a `testContract` with a literal
`MurabahaParams`. Marshal it to RawMessage:

```go
func testContract(id string) sharia.Contract {
	raw, _ := json.Marshal(sharia.MurabahaParams{
		AssetCode: "VHCL42A",
		Cost:      sharia.Monetary{Asset: "SAR2", Amount: 10000000},
		Markup:    sharia.Monetary{Asset: "SAR2", Amount: 1000000},
		Client:    "@client:ameziane", Supplier: "@supplier:toyota",
		BankTreasury: "@bank:treasury", Installments: 24,
		FirstDue: "2026-07-01T00:00:00Z", PeriodDays: 30,
	})
	return sharia.Contract{
		ID: id, Type: sharia.TypeMurabaha, State: sharia.StatePromise,
		Params: raw, TemplateVersion: sharia.TemplateVersionMurabaha,
		CreatedAt: "2026-06-10T00:00:00Z", UpdatedAt: "2026-06-10T00:00:00Z",
	}
}
```

These storage tests assert round-trip persistence, not engine behavior;
adapting their fixture to the new field type is allowed (the A–G *engine*
tests stay untouched). Add `encoding/json` import where needed.

- [ ] **Step 5: Make engine.go compile against RawMessage** — in `engine.go`,
`acquire/sell/payInstallment/earlySettle/latePenalty/cancel` read `c.Params.X`.
Temporarily, at the top of each, decode once:

```go
var p MurabahaParams
_ = json.Unmarshal(c.Params, &p)
```

and replace `c.Params.X` with `p.X` in those methods. In `Create`, marshal the
incoming params: the request still carries `MurabahaParams` for now (Task 4
generalizes `CreateRequest`). This step is purely to restore compilation with
identical behavior.

- [ ] **Step 6: Verify all green**

Run: `go test ./... -count=1`
Expected: every package `ok`, A–G unchanged and passing.

- [ ] **Step 7: Commit**

```bash
git add sharia/types.go sharia/engine.go storage/
git commit -m "refactor(sharia): Contract.Params is json.RawMessage (behavior-preserving)"
```

---

### Task 3: Implement murabahaKind by extracting engine logic

**Files:** Modify `sharia/murabaha.go` (add the kind), `sharia/engine.go` (will call it in Task 4)

`murabaha.go` already has the pure posting builders (`AcquirePostings`,
`SellPostings`, `PayInstallmentPostings`, `EarlySettlePostings`,
`PenaltyPostings`, `CancelPostings`) and `TransitionAllowed`. Wrap them in a kind.

- [ ] **Step 1: Append `murabahaKind` to `sharia/murabaha.go`**

```go
type murabahaKind struct{}

func init() { register(murabahaKind{}) }

func (murabahaKind) Type() string { return TypeMurabaha }

func (murabahaKind) DecodeParams(raw json.RawMessage) (Params, error) {
	var p MurabahaParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, newError(ErrInvalidParams, "invalid murabaha params: "+err.Error())
	}
	return p, nil
}

func (murabahaKind) ValidateParams(p Params) error {
	mp := p.(MurabahaParams)
	return (&mp).Validate()
}

func (murabahaKind) BuildSchedule(p Params) ([]Installment, error) {
	mp := p.(MurabahaParams)
	return BuildSchedule(mp.Cost.Amount, mp.Markup.Amount, mp.Installments, mp.FirstDue, mp.PeriodDays)
}

func (murabahaKind) AllowedTransitions(from string) []string {
	out := []string{}
	for name := range murabahaFSM[from] {
		out = append(out, name)
	}
	return out
}

// ShariaGate preserves today's behavior: selling a non-possessed asset is
// SS-8, checked before the FSM even from PROMISE.
func (murabahaKind) ShariaGate(led LedgerPort, c Contract, p Params, ev Event) error {
	if ev.Name != TransitionSell {
		return nil
	}
	mp := p.(MurabahaParams)
	inv, err := led.GetAccount(InventoryAccount(c.ID))
	if err != nil {
		return err
	}
	if inv.Balances[mp.AssetCode] < 1 {
		return &Error{Code: ErrShariaViolation, Message: "sale of non-possessed asset", StandardRef: RefSS8}
	}
	return nil
}
```

(Note: `murabahaFSM`, `TransitionAllowed`, the posting builders and the
`MurabahaParams.Validate` method already exist in the repo. `register` is from
Task 1. `init()` self-registers the kind.)

- [ ] **Step 2: Add `Preconditions` and `BuildPlan` to `murabaha.go`** — these
move the bodies of the engine's `acquire/sell/payInstallment/earlySettle/
latePenalty/cancel` methods, split into a precondition phase (balance/seq checks
that produce `deny`-able errors) and a plan phase (postings + state + marks).
Reproduce the EXACT logic currently in `engine.go` for each transition; the
only change is *where* it lives. Each transition's precondition error codes and
posting amounts must be byte-identical to today (A–G proves this).

Concretely, `BuildPlan` switches on `ev.Name`:
- `acquire`: postings `AcquirePostings`, ref `<id>:acquire`, newState ACQUIRED, ref SS-8.
- `sell`: postings `SellPostings`, ref `<id>:sell`, newState SOLD, ref SS-8.
- `pay_installment`: resolve target installment (input.Seq or smallest unpaid),
  postings `PayInstallmentPostings`, ref `<id>:pay:<seq>`, mark seq paid; if it
  was the last unpaid → newState SETTLED + `ExtraAudit` settled event; else SOLD.
- `early_settle`: compute rest_total/rest_profit, rebate default = rest_profit,
  postings `EarlySettlePostings`, ref `<id>:early_settle`, mark all unpaid
  settled_early, newState SETTLED + settled ExtraAudit.
- `late_penalty`: `PenaltyPostings` (charity gate inside), ref
  `<id>:late_penalty[:seq]`, newState unchanged, event Penalty, ref SS-3.
- `cancel`: `CancelPostings`; PROMISE → no postings (engine handles the
  no-posting path), ACQUIRED → inventory move; newState CANCELLED.

`Preconditions` does the balance/existence checks that currently precede the
commit (treasury ≥ cost for acquire; client ≥ amount for pay/early; installment
exists & unpaid; etc.), returning the same typed errors.

- [ ] **Step 3: Build (kind not yet called by engine)**

Run: `go build ./...`
Expected: compiles. Tests still pass because the engine still uses its own
switch (Task 4 swaps it).

- [ ] **Step 4: Commit**

```bash
git add sharia/murabaha.go
git commit -m "refactor(sharia): murabahaKind implements ContractKind (extracted, unwired)"
```

---

### Task 4: Wire the engine to the registry generically

**Files:** Modify `sharia/engine.go`, generalize `CreateRequest`

- [ ] **Step 1: Generalize `CreateRequest`** in `engine.go`

```go
type CreateRequest struct {
	Type   string          `json:"type"`
	ID     string          `json:"id"`
	Params json.RawMessage `json:"params"`
}
```

- [ ] **Step 2: Rewrite `Engine.Create`** to route via the registry

```go
func (e *Engine) Create(req CreateRequest) (Contract, []Installment, error) {
	var c Contract
	if req.Type == "" {
		req.Type = TypeMurabaha
	}
	kind, ok := kindFor(req.Type)
	if !ok {
		return c, nil, newError(ErrInvalidParams, "unknown contract type "+req.Type)
	}
	if req.ID == "" {
		req.ID = GenerateContractID()
	}
	if !ContractIDIsValid(req.ID) {
		return c, nil, newError(ErrInvalidParams, "id must match ^[a-z][a-z0-9_]{3,40}$")
	}
	p, err := kind.DecodeParams(req.Params)
	if err != nil {
		return c, nil, err
	}
	if err := kind.ValidateParams(p); err != nil {
		return c, nil, err
	}
	if _, err := e.store.GetContract(req.ID); err == nil {
		return c, nil, &Error{Code: ErrDuplicate, Message: "contract already exists", ContractID: req.ID}
	}
	schedule, err := kind.BuildSchedule(p)
	if err != nil {
		return c, nil, newError(ErrInvalidParams, err.Error())
	}
	ts := now()
	// re-marshal the decoded+defaulted params so stored params are canonical
	raw, _ := json.Marshal(p)
	c = Contract{
		ID: req.ID, Type: req.Type, State: kindInitialState(req.Type),
		Params: raw, TemplateVersion: templateVersionFor(req.Type),
		CreatedAt: ts, UpdatedAt: ts,
	}
	if err := e.store.SaveContract(c); err != nil {
		if isDuplicateErr(err) {
			return c, nil, &Error{Code: ErrDuplicate, Message: "contract already exists", ContractID: req.ID}
		}
		return c, nil, err
	}
	if err := e.store.SaveSchedule(c.ID, schedule); err != nil {
		return c, nil, err
	}
	if _, err := e.appendAudit(AuditEvent{
		ContractID: c.ID, Event: EventCreated, Decision: DecisionAllowed,
		StandardRef: createStandardRef(req.Type), TxID: -1, Payload: string(raw), CreatedAt: ts,
	}); err != nil {
		return c, nil, err
	}
	return c, schedule, nil
}
```

Add helpers in `kind.go`: `kindInitialState`, `templateVersionFor`,
`createStandardRef` — small maps keyed by type (murabaha→PROMISE/`murabaha/1.0.0`/SS-8;
ijarah→PROMISE/`ijarah/1.0.0`/SS-9). For Task 4 only murabaha entries exist.

- [ ] **Step 3: Rewrite `Engine.Transition`** to be generic

```go
func (e *Engine) Transition(contractID, name string, input TransitionInput) (TransitionResult, error) {
	var res TransitionResult
	lock := e.km.get(contractID)
	lock.Lock()
	defer lock.Unlock()

	c, err := e.store.GetContract(contractID)
	if err != nil {
		return res, &Error{Code: ErrNotFound, Message: "contract not found", ContractID: contractID}
	}
	kind, ok := kindFor(c.Type)
	if !ok {
		return res, &Error{Code: ErrInvalidParams, Message: "unknown contract type " + c.Type, ContractID: contractID}
	}
	p, err := kind.DecodeParams(c.Params)
	if err != nil {
		return res, err
	}
	ev := Event{Name: name, Input: input}

	// 1. Sharia gate (pre-FSM priority)
	if gErr := kind.ShariaGate(e.ledger, c, p, ev); gErr != nil {
		if se, ok := gErr.(*Error); ok {
			return res, e.deny(contractID, name, se)
		}
		return res, gErr
	}
	// 2. FSM
	if !contains(kind.AllowedTransitions(c.State), name) {
		return res, e.deny(contractID, name, &Error{
			Code: ErrInvalidTransition,
			Message: fmt.Sprintf("transition %q is not allowed from state %s", name, c.State),
		})
	}
	// 3. load schedule + preconditions
	sched, err := e.store.GetSchedule(contractID)
	if err != nil {
		return res, err
	}
	if pErr := kind.Preconditions(e.ledger, c, p, sched, ev); pErr != nil {
		if se, ok := pErr.(*Error); ok {
			return res, e.deny(contractID, name, se)
		}
		return res, pErr
	}
	// 4. build plan
	plan, err := kind.BuildPlan(e.ledger, c, p, sched, ev)
	if err != nil {
		if se, ok := err.(*Error); ok {
			return res, e.deny(contractID, name, se)
		}
		return res, err
	}
	// 5. commit + generic post-commit (see Task 5)
	return e.applyPlan(c, name, plan)
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Delete the old per-transition methods** from engine.go
(`acquire`, `sell`, `payInstallment`, `earlySettle`, `latePenalty`, `cancel`,
`commitTransition`) now that their logic lives in `murabahaKind` + `applyPlan`.

- [ ] **Step 5: Verify — the whole point**

Run: `go test ./... -count=1`
Expected: A–G and all sharia/api/storage tests `ok`, UNCHANGED. If any fail,
the extraction diverged from the original behavior — fix the kind, not the test.

- [ ] **Step 6: Commit**

```bash
git add sharia/engine.go sharia/kind.go
git commit -m "refactor(sharia): engine routes transitions through the registry"
```

---

### Task 5: Generic plan application (applyPlan)

**Files:** Modify `sharia/engine.go`

- [ ] **Step 1: Add `applyPlan`** — the generic commit + post-commit, replacing
the old `commitTransition`. Handles: no-posting transitions (cancel from
PROMISE), commit, SaveMeta, installment marks, state update, audit, extra audit.

```go
func (e *Engine) applyPlan(c Contract, name string, plan TransitionPlan) (TransitionResult, error) {
	var res TransitionResult
	ts := now()
	txID := int64(-1)

	if len(plan.Postings) > 0 {
		tx := core.Transaction{Postings: plan.Postings, Reference: plan.Reference}
		committed, err := e.ledger.Commit([]core.Transaction{tx})
		if err != nil {
			if isDuplicateErr(err) {
				return res, e.deny(c.ID, name, &Error{Code: ErrDuplicate,
					Message: fmt.Sprintf("transition already committed (reference %q)", plan.Reference)})
			}
			if strings.HasPrefix(err.Error(), "balance.insufficient") {
				return res, e.deny(c.ID, name, &Error{Code: ErrPrecondition, Message: err.Error()})
			}
			return res, err
		}
		txID = committed[0].ID
		e.ledger.SaveMeta("transaction", fmt.Sprint(txID), core.Metadata{
			"sharia/contract": json.RawMessage(fmt.Sprintf("%q", c.ID)),
			"sharia/event":    json.RawMessage(fmt.Sprintf("%q", name)),
			"sharia/ref":      json.RawMessage(fmt.Sprintf("%q", plan.StandardRef)),
		})
	}

	for _, m := range plan.Marks {
		if err := e.store.MarkInstallment(c.ID, m.Seq, m.Status, txID, ts); err != nil {
			return res, err
		}
	}
	if plan.NewState != "" && plan.NewState != c.State {
		if err := e.store.UpdateContractState(c.ID, plan.NewState, ts); err != nil {
			return res, err
		}
	}
	event := plan.Event
	if event == "" {
		event = EventTransition
	}
	if _, err := e.appendAudit(AuditEvent{
		ContractID: c.ID, Event: event, Transition: name, Decision: DecisionAllowed,
		StandardRef: plan.StandardRef, TxID: txID, Payload: plan.Payload,
	}); err != nil {
		return res, err
	}
	for _, ex := range plan.ExtraAudit {
		ex.ContractID = c.ID
		if _, err := e.appendAudit(ex); err != nil {
			return res, err
		}
	}

	newState := plan.NewState
	if newState == "" {
		newState = c.State
	}
	c.State, c.UpdatedAt = newState, ts
	return TransitionResult{Contract: c, Transition: name, TxID: txID, NewState: newState}, nil
}
```

(The murabaha pay_installment's "settled when last paid" logic moves into
`murabahaKind.BuildPlan`: it inspects `sched`, decides newState + whether to add
a settled `ExtraAudit`. This keeps `applyPlan` contract-agnostic.)

- [ ] **Step 2: Verify green, commit**

Run: `go test ./... -count=1` → all `ok`, A–G unchanged.
```bash
git add sharia/engine.go && git commit -m "refactor(sharia): generic applyPlan post-commit"
```

---

## PART 2 — Ijarah (operating lease), full TDD

### Task 6: Ijarah params + validation

**Files:** Modify `sharia/types.go`; Test `sharia/ijarah_test.go`

- [ ] **Step 1: Write failing test** in `sharia/ijarah_test.go`

```go
package sharia

import "testing"

func validIjarah() IjarahParams {
	return IjarahParams{
		AssetCode: "VHCL1", Cost: Monetary{Asset: "DZD.2", Amount: 10000000},
		Rent: Monetary{Asset: "DZD.2", Amount: 500000}, Client: "@client:anis",
		Supplier: "@supplier:toyota", BankTreasury: "@bank:treasury",
		Periods: 24, FirstDue: "2026-07-01T00:00:00Z", PeriodDays: 30,
	}
}

func TestIjarahValidateOK(t *testing.T) {
	p := validIjarah()
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestIjarahValidateDefaults(t *testing.T) {
	p := validIjarah()
	p.BankTreasury, p.PeriodDays = "", 0
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.BankTreasury != "@bank:treasury" || p.PeriodDays != 30 {
		t.Fatalf("defaults not applied: %+v", p)
	}
}

func TestIjarahValidateRejects(t *testing.T) {
	cases := []struct{ name, detail string; mut func(*IjarahParams) }{
		{"zero cost", "cost", func(p *IjarahParams) { p.Cost.Amount = 0 }},
		{"zero rent", "rent", func(p *IjarahParams) { p.Rent.Amount = 0 }},
		{"asset mismatch", "asset", func(p *IjarahParams) { p.Rent.Asset = "EUR.2" }},
		{"asset_code equals money", "asset_code", func(p *IjarahParams) { p.AssetCode = "DZD" ; p.Cost.Asset="DZD"; p.Rent.Asset="DZD" }},
		{"zero periods", "periods", func(p *IjarahParams) { p.Periods = 0 }},
		{"rent under cost", "exceed cost", func(p *IjarahParams) { p.Rent.Amount = 100; p.Periods = 2 }},
		{"bad client", "client", func(p *IjarahParams) { p.Client = "anis" }},
		{"bad due", "first_due", func(p *IjarahParams) { p.FirstDue = "nope" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := validIjarah()
			c.mut(&p)
			err := p.Validate()
			if err == nil {
				t.Fatal("expected error")
			}
			se, ok := err.(*Error)
			if !ok || se.Code != ErrInvalidParams {
				t.Fatalf("expected ERR_INVALID_PARAMS, got %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run** `go test ./sharia/ -run TestIjarahValidate` → FAIL (undefined IjarahParams)

- [ ] **Step 3: Add to `sharia/types.go`**

```go
const (
	TypeIjarah = "ijarah"

	StateLeased    = "LEASED"
	StateCompleted = "COMPLETED"

	TemplateVersionIjarah = "ijarah/1.0.0"
	RefSS9                = "AAOIFI-SS-9"
)

type IjarahParams struct {
	AssetCode    string   `json:"asset_code"`
	Cost         Monetary `json:"cost"`
	Rent         Monetary `json:"rent"`
	Client       string   `json:"client"`
	Supplier     string   `json:"supplier"`
	BankTreasury string   `json:"bank_treasury"`
	Periods      int      `json:"periods"`
	FirstDue     string   `json:"first_due"`
	PeriodDays   int      `json:"period_days"`
}

func (p *IjarahParams) Validate() error {
	if p.BankTreasury == "" {
		p.BankTreasury = "@bank:treasury"
	}
	if p.PeriodDays == 0 {
		p.PeriodDays = 30
	}
	switch {
	case p.Cost.Amount <= 0:
		return newError(ErrInvalidParams, "cost.amount must be > 0")
	case p.Rent.Amount <= 0:
		return newError(ErrInvalidParams, "rent.amount must be > 0")
	case !assetRe.MatchString(p.Cost.Asset):
		return newError(ErrInvalidParams, "cost.asset is not a valid asset code")
	case p.Cost.Asset != p.Rent.Asset:
		return newError(ErrInvalidParams, "cost and rent asset must match")
	case !assetRe.MatchString(p.AssetCode):
		return newError(ErrInvalidParams, "asset_code is not a valid asset code")
	case p.AssetCode == p.Cost.Asset:
		return newError(ErrInvalidParams, "asset_code must differ from the monetary asset")
	case p.Periods < 1:
		return newError(ErrInvalidParams, "periods must be >= 1")
	case !accountRe.MatchString(p.Client):
		return newError(ErrInvalidParams, "client is not a valid account address")
	case !accountRe.MatchString(p.Supplier):
		return newError(ErrInvalidParams, "supplier is not a valid account address")
	case !accountRe.MatchString(p.BankTreasury):
		return newError(ErrInvalidParams, "bank_treasury is not a valid account address")
	case p.PeriodDays < 1:
		return newError(ErrInvalidParams, "period_days must be >= 1")
	}
	if _, err := time.Parse(time.RFC3339, p.FirstDue); err != nil {
		return newError(ErrInvalidParams, "first_due is not a valid RFC3339 date")
	}
	if p.Rent.Amount*int64(p.Periods) <= p.Cost.Amount {
		return newError(ErrInvalidParams, "rent over the term must exceed cost")
	}
	return nil
}

// Ijarah account helpers
func AssetAccount(id string) string      { return "@contracts:" + id + ":asset" }
func InUseAccount(client string) string  { return client + ":in_use" }

const (
	IjarahIncomeAccount = "@bank:income:ijarah"
	DepreciationAccount = "@bank:expense:depreciation"
	ReturnedInventory   = "@bank:inventory:returned"
)
```

- [ ] **Step 4: Run** `go test ./sharia/ -run TestIjarahValidate` → PASS
- [ ] **Step 5: Commit** `git add sharia/types.go sharia/ijarah_test.go && git commit -m "feat(sharia): ijarah params and validation"`

---

### Task 7: Ijarah schedule (fixed rent + straight-line depreciation)

**Files:** Modify `sharia/schedule.go`, `sharia/types.go` (Installment note); Test `sharia/ijarah_test.go`

The `Installment` struct is reused. For ijarah rows: `Amount` = rent,
`ProfitPart` = 0, `PrincipalPart` = 0, and a new field carries depreciation.
Add `DepreciationPart int64` to `Installment` (json `depreciation_part,omitempty`)
— additive, murabaha leaves it 0. Persist it (Task 8).

- [ ] **Step 1: Failing test**

```go
func TestBuildIjarahScheduleReference(t *testing.T) {
	items, err := BuildIjarahSchedule(10000000, 500000, 24, "2026-07-01T00:00:00Z", 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 24 {
		t.Fatalf("expected 24, got %d", len(items))
	}
	var rent, depr int64
	for i, it := range items {
		if it.Amount != 500000 {
			t.Fatalf("seq %d rent: expected 500000, got %d", it.Seq, it.Amount)
		}
		if it.Status != StatusPending {
			t.Fatalf("seq %d: expected pending", it.Seq)
		}
		rent += it.Amount
		depr += it.DepreciationPart
		_ = i
	}
	if rent != 12000000 {
		t.Fatalf("total rent: expected 12000000, got %d", rent)
	}
	if depr != 10000000 {
		t.Fatalf("total depreciation: expected 10000000, got %d", depr)
	}
	for i := 0; i < 23; i++ {
		if items[i].DepreciationPart != 416666 {
			t.Fatalf("seq %d depr: expected 416666, got %d", i+1, items[i].DepreciationPart)
		}
	}
	if items[23].DepreciationPart != 416682 {
		t.Fatalf("seq 24 depr: expected 416682, got %d", items[23].DepreciationPart)
	}
}
```

- [ ] **Step 2: Run** → FAIL (undefined BuildIjarahSchedule)

- [ ] **Step 3: Implement** in `sharia/schedule.go`

```go
// BuildIjarahSchedule: fixed rent per period, straight-line depreciation of
// cost over the term (remainder on the last period). Integer-exact (I-6).
func BuildIjarahSchedule(cost, rent int64, periods int, firstDue string, periodDays int) ([]Installment, error) {
	first, err := time.Parse(time.RFC3339, firstDue)
	if err != nil {
		return nil, fmt.Errorf("invalid first_due date: %w", err)
	}
	depr := SplitEven(cost, periods)
	items := make([]Installment, periods)
	for i := 0; i < periods; i++ {
		items[i] = Installment{
			Seq:              i + 1,
			DueDate:          first.UTC().AddDate(0, 0, i*periodDays).Format(time.RFC3339),
			Amount:           rent,
			DepreciationPart: depr[i],
			Status:           StatusPending,
			PaidTxID:         -1,
		}
	}
	return items, nil
}
```

Add `DepreciationPart int64 \`json:"depreciation_part,omitempty"\`` to the
`Installment` struct in `types.go`.

- [ ] **Step 4: Run** `go test ./sharia/ -run TestBuildIjarahSchedule` → PASS
- [ ] **Step 5: Commit** `git add sharia/schedule.go sharia/types.go && git commit -m "feat(sharia): ijarah schedule (fixed rent + straight-line depreciation)"`

---

### Task 8: Persist DepreciationPart

**Files:** Modify `storage/sqlite/migration/v002.sql` is frozen — add `v003.sql`;
`storage/postgres/migration/v003.sql`; `storage/sqlite/sharia.go`,
`storage/postgres/sharia.go`

- [ ] **Step 1:** Create `storage/sqlite/migration/v003.sql`:

```sql
--statement
ALTER TABLE sharia_schedule ADD COLUMN "depreciation_part" integer DEFAULT 0;
```

and `storage/postgres/migration/v003.sql`:

```sql
--statement
ALTER TABLE "VAR_LEDGER_NAME".sharia_schedule ADD COLUMN "depreciation_part" bigint DEFAULT 0;
```

(Migrations run in filename order; `ADD COLUMN` is idempotent-safe enough for
fresh DBs. For DBs already migrated, the column is added once. The legacy-repair
from pilot-hardening renames incompatible tables aside, so this runs on our schema.)

- [ ] **Step 2:** In both `SaveSchedule` and `GetSchedule` (sqlite + postgres),
add the `depreciation_part` column to the insert cols/values and the select
list, scanning into `it.DepreciationPart`. Follow the existing column pattern.

- [ ] **Step 3: Test** — extend `storage/sqlite/sharia_test.go` `TestScheduleRoundTrip`
to set and assert a `DepreciationPart` on one item.

```go
// after building items, set depreciation on seq 1 and assert it round-trips
items[0].DepreciationPart = 416666
// ... SaveSchedule, GetSchedule ...
if got[0].DepreciationPart != 416666 {
	t.Fatalf("depreciation_part not persisted: %d", got[0].DepreciationPart)
}
```

- [ ] **Step 4: Run** `go test ./storage/sqlite/ -count=1` → PASS
- [ ] **Step 5: Commit** `git add storage/ && git commit -m "feat(storage): persist installment depreciation_part (migration v003)"`

---

### Task 9: Ijarah posting builders

**Files:** Create `sharia/ijarah.go`; Test `sharia/ijarah_test.go`

- [ ] **Step 1: Failing tests** (pure functions, mirror murabaha_test patterns)

```go
func TestIjarahAcquirePostings(t *testing.T) {
	p := validIjarah()
	got := IjarahAcquirePostings("ijr1", p)
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d", len(got))
	}
	assertPosting(t, got[0], "@bank:treasury", "@supplier:toyota", "DZD.2", 10000000)
	assertPosting(t, got[1], "@world", "@contracts:ijr1:asset", "VHCL1", 1)
	assertPosting(t, got[2], "@world", "@contracts:ijr1:asset", "DZD.2", 10000000)
}

func TestIjarahLeasePostings(t *testing.T) {
	p := validIjarah()
	got := IjarahLeasePostings("ijr1", p)
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	assertPosting(t, got[0], "@contracts:ijr1:asset", "@client:anis:in_use", "VHCL1", 1)
}

func TestIjarahPayRentPostings(t *testing.T) {
	p := validIjarah()
	inst := Installment{Seq: 1, Amount: 500000, DepreciationPart: 416666}
	got := IjarahPayRentPostings("ijr1", p, inst, false)
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d", len(got))
	}
	assertPosting(t, got[0], "@client:anis", "@bank:treasury", "DZD.2", 500000)
	assertPosting(t, got[1], "@world", "@bank:income:ijarah", "DZD.2", 500000)
	assertPosting(t, got[2], "@contracts:ijr1:asset", "@bank:expense:depreciation", "DZD.2", 416666)
}

func TestIjarahPayRentPostingsLastReturnsAsset(t *testing.T) {
	p := validIjarah()
	inst := Installment{Seq: 24, Amount: 500000, DepreciationPart: 416682}
	got := IjarahPayRentPostings("ijr1", p, inst, true) // last period
	if len(got) != 4 {
		t.Fatalf("expected 4 on last period, got %d", len(got))
	}
	assertPosting(t, got[3], "@client:anis:in_use", "@bank:inventory:returned", "VHCL1", 1)
}

func TestIjarahPenaltyRejectsNonCharity(t *testing.T) {
	p := validIjarah()
	if _, err := IjarahPenaltyPostings("ijr1", p, 20000, "@bank:income:ijarah"); err == nil {
		t.Fatal("expected SS-3 violation")
	}
	got, err := IjarahPenaltyPostings("ijr1", p, 20000, "@charity:pool")
	if err != nil || len(got) != 1 {
		t.Fatalf("charity penalty should pass: %v", err)
	}
	assertPosting(t, got[0], "@client:anis", "@charity:pool", "DZD.2", 20000)
}

func TestIjarahCancelFromAcquired(t *testing.T) {
	p := validIjarah()
	got := IjarahCancelPostings("ijr1", p, StateAcquired)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	assertPosting(t, got[0], "@contracts:ijr1:asset", "@bank:inventory:unsold", "VHCL1", 1)
	assertPosting(t, got[1], "@contracts:ijr1:asset", "@world", "DZD.2", 10000000)
}
```

- [ ] **Step 2: Run** → FAIL (undefined)

- [ ] **Step 3: Implement `sharia/ijarah.go`** (FSM map + posting builders)

```go
package sharia

import (
	"strings"

	"github.com/amezianechayer/corren/core"
)

const (
	TransitionLease   = "lease"
	TransitionPayRent = "pay_rent"
)

var ijarahFSM = map[string]map[string]bool{
	StatePromise:  {TransitionAcquire: true, TransitionCancel: true},
	StateAcquired: {TransitionLease: true, TransitionCancel: true},
	StateLeased:   {TransitionPayRent: true, TransitionLatePenalty: true},
}

func IjarahAcquirePostings(id string, p IjarahParams) []core.Posting {
	return []core.Posting{
		{Source: p.BankTreasury, Destination: p.Supplier, Asset: p.Cost.Asset, Amount: p.Cost.Amount},
		{Source: core.WORLD, Destination: AssetAccount(id), Asset: p.AssetCode, Amount: 1},
		{Source: core.WORLD, Destination: AssetAccount(id), Asset: p.Cost.Asset, Amount: p.Cost.Amount},
	}
}

func IjarahLeasePostings(id string, p IjarahParams) []core.Posting {
	return []core.Posting{
		{Source: AssetAccount(id), Destination: InUseAccount(p.Client), Asset: p.AssetCode, Amount: 1},
	}
}

func IjarahPayRentPostings(id string, p IjarahParams, inst Installment, last bool) []core.Posting {
	out := []core.Posting{
		{Source: p.Client, Destination: p.BankTreasury, Asset: p.Cost.Asset, Amount: inst.Amount},
		{Source: core.WORLD, Destination: IjarahIncomeAccount, Asset: p.Cost.Asset, Amount: inst.Amount},
		{Source: AssetAccount(id), Destination: DepreciationAccount, Asset: p.Cost.Asset, Amount: inst.DepreciationPart},
	}
	if last {
		out = append(out, core.Posting{
			Source: InUseAccount(p.Client), Destination: ReturnedInventory, Asset: p.AssetCode, Amount: 1,
		})
	}
	return out
}

func IjarahPenaltyPostings(id string, p IjarahParams, amount int64, destination string) ([]core.Posting, error) {
	if !strings.HasPrefix(destination, CharityPrefix) {
		return nil, &Error{Code: ErrShariaViolation,
			Message: "late penalty destination must be a @charity: account", StandardRef: RefSS3, ContractID: id}
	}
	return []core.Posting{
		{Source: p.Client, Destination: destination, Asset: p.Cost.Asset, Amount: amount},
	}, nil
}

func IjarahCancelPostings(id string, p IjarahParams, fromState string) []core.Posting {
	if fromState != StateAcquired {
		return nil
	}
	return []core.Posting{
		{Source: AssetAccount(id), Destination: UnsoldInventory, Asset: p.AssetCode, Amount: 1},
		{Source: AssetAccount(id), Destination: core.WORLD, Asset: p.Cost.Asset, Amount: p.Cost.Amount},
	}
}
```

- [ ] **Step 4: Run** `go test ./sharia/ -run TestIjarah` → PASS
- [ ] **Step 5: Commit** `git add sharia/ijarah.go sharia/ijarah_test.go && git commit -m "feat(sharia): ijarah posting builders"`

---

### Task 10: ijarahKind (wire Ijarah into the engine)

**Files:** Modify `sharia/ijarah.go`, `sharia/kind.go` (type maps)

- [ ] **Step 1: Append `ijarahKind` to `sharia/ijarah.go`**

```go
type ijarahKind struct{}

func init() { register(ijarahKind{}) }

func (ijarahKind) Type() string { return TypeIjarah }

func (ijarahKind) DecodeParams(raw json.RawMessage) (Params, error) {
	var p IjarahParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, newError(ErrInvalidParams, "invalid ijarah params: "+err.Error())
	}
	return p, nil
}

func (ijarahKind) ValidateParams(p Params) error {
	ip := p.(IjarahParams)
	return (&ip).Validate()
}

func (ijarahKind) BuildSchedule(p Params) ([]Installment, error) {
	ip := p.(IjarahParams)
	return BuildIjarahSchedule(ip.Cost.Amount, ip.Rent.Amount, ip.Periods, ip.FirstDue, ip.PeriodDays)
}

func (ijarahKind) AllowedTransitions(from string) []string {
	out := []string{}
	for name := range ijarahFSM[from] {
		out = append(out, name)
	}
	return out
}

// ShariaGate: leasing a non-owned asset is SS-9, before the FSM.
func (ijarahKind) ShariaGate(led LedgerPort, c Contract, p Params, ev Event) error {
	if ev.Name != TransitionLease {
		return nil
	}
	ip := p.(IjarahParams)
	asset, err := led.GetAccount(AssetAccount(c.ID))
	if err != nil {
		return err
	}
	if asset.Balances[ip.AssetCode] < 1 {
		return &Error{Code: ErrShariaViolation, Message: "lease of non-owned asset", StandardRef: RefSS9}
	}
	return nil
}

func (ijarahKind) Preconditions(led LedgerPort, c Contract, p Params, sched []Installment, ev Event) error {
	ip := p.(IjarahParams)
	switch ev.Name {
	case TransitionAcquire:
		t, err := led.GetAccount(ip.BankTreasury)
		if err != nil {
			return err
		}
		if t.Balances[ip.Cost.Asset] < ip.Cost.Amount {
			return &Error{Code: ErrPrecondition, Message: "insufficient treasury balance to acquire the asset"}
		}
	case TransitionPayRent:
		inst := ijarahTarget(sched, ev.Input.Seq)
		if inst == nil {
			return &Error{Code: ErrNotFound, Message: "no such installment"}
		}
		if inst.Status == StatusPaid {
			return &Error{Code: ErrDuplicate, Message: "installment already paid"}
		}
		cl, err := led.GetAccount(ip.Client)
		if err != nil {
			return err
		}
		if cl.Balances[ip.Cost.Asset] < inst.Amount {
			return &Error{Code: ErrPrecondition, Message: "insufficient client balance"}
		}
	case TransitionLatePenalty:
		if ev.Input.Amount <= 0 {
			return &Error{Code: ErrInvalidParams, Message: "penalty amount must be > 0"}
		}
	}
	return nil
}

func (ijarahKind) BuildPlan(led LedgerPort, c Contract, p Params, sched []Installment, ev Event) (TransitionPlan, error) {
	ip := p.(IjarahParams)
	switch ev.Name {
	case TransitionAcquire:
		payload, _ := json.Marshal(map[string]interface{}{"cost": ip.Cost, "supplier": ip.Supplier})
		return TransitionPlan{
			Postings: IjarahAcquirePostings(c.ID, ip), Reference: c.ID + ":acquire",
			NewState: StateAcquired, StandardRef: RefSS9, Payload: string(payload),
		}, nil
	case TransitionLease:
		return TransitionPlan{
			Postings: IjarahLeasePostings(c.ID, ip), Reference: c.ID + ":lease",
			NewState: StateLeased, StandardRef: RefSS9, Payload: "{}",
		}, nil
	case TransitionPayRent:
		inst := ijarahTarget(sched, ev.Input.Seq)
		last := ijarahIsLastUnpaid(sched, inst.Seq)
		newState := StateLeased
		var extra []AuditEvent
		if last {
			newState = StateCompleted
			extra = []AuditEvent{{Event: EventSettled, Decision: DecisionAllowed,
				StandardRef: "AAOIFI-FAS-32 (simplified v1)", TxID: -1, Payload: "{}"}}
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"seq": inst.Seq, "rent": inst.Amount, "depreciation_part": inst.DepreciationPart,
			"accounting": "FAS 32 simplified (v1); depreciation basis to be validated by a Sharia/accounting advisor",
		})
		return TransitionPlan{
			Postings: IjarahPayRentPostings(c.ID, ip, *inst, last),
			Reference: fmt.Sprintf("%s:rent:%d", c.ID, inst.Seq),
			NewState: newState, StandardRef: RefFAS28, // see note Step 2
			Payload: string(payload),
			Marks:   []InstallmentMark{{Seq: inst.Seq, Status: StatusPaid}},
			ExtraAudit: extra,
		}, nil
	case TransitionLatePenalty:
		dest := ev.Input.Destination
		if dest == "" {
			dest = DefaultCharityPool
		}
		postings, err := IjarahPenaltyPostings(c.ID, ip, ev.Input.Amount, dest)
		if err != nil {
			return TransitionPlan{}, err
		}
		ref := c.ID + ":late_penalty"
		if ev.Input.Seq > 0 {
			ref = fmt.Sprintf("%s:late_penalty:%d", c.ID, ev.Input.Seq)
		}
		payload, _ := json.Marshal(map[string]interface{}{"seq": ev.Input.Seq, "amount": ev.Input.Amount, "destination": dest})
		return TransitionPlan{
			Postings: postings, Reference: ref, NewState: c.State,
			StandardRef: RefSS3, Event: EventPenalty, Payload: string(payload),
		}, nil
	case TransitionCancel:
		return TransitionPlan{
			Postings: IjarahCancelPostings(c.ID, ip, c.State), Reference: c.ID + ":cancel",
			NewState: StateCancelled, Payload: "{}",
		}, nil
	}
	return TransitionPlan{}, &Error{Code: ErrInvalidTransition, Message: "unknown transition " + ev.Name}
}

func ijarahTarget(sched []Installment, seq int) *Installment {
	if seq > 0 {
		for i := range sched {
			if sched[i].Seq == seq {
				return &sched[i]
			}
		}
		return nil
	}
	for i := range sched {
		if sched[i].Status == StatusPending || sched[i].Status == StatusOverdue {
			return &sched[i]
		}
	}
	return nil
}

func ijarahIsLastUnpaid(sched []Installment, seq int) bool {
	for _, it := range sched {
		if it.Seq != seq && (it.Status == StatusPending || it.Status == StatusOverdue) {
			return false
		}
	}
	return true
}
```

- [ ] **Step 2: Add a dedicated FAS-32 ref + type maps** — in `types.go` add
`RefFAS32 = "AAOIFI-FAS-32 (simplified v1)"` and use it in the pay_rent plan
(`StandardRef: RefFAS32`) and the settled ExtraAudit. In `kind.go` extend the
helper maps: `kindInitialState[ijarah]=PROMISE`, `templateVersionFor[ijarah]=TemplateVersionIjarah`, `createStandardRef[ijarah]=RefSS9`. Add `TransitionCancel` to `ijarahFSM` reachable states already present (PROMISE, ACQUIRED).

- [ ] **Step 3: Build** `go build ./...` → compiles, ijarahKind self-registers.
- [ ] **Step 4: Run** full suite `go test ./... -count=1` → all green (murabaha untouched).
- [ ] **Step 5: Commit** `git add sharia/ && git commit -m "feat(sharia): ijarahKind wired into the multi-contract engine"`

---

### Task 11: Ijarah end-to-end scenario (the proof)

**Files:** Test `sharia/ijarah_test.go` (uses the `withEngine`/`fund`/`assertBal` helpers already in `engine_test.go`, same `sharia_test` package — note: ijarah_test.go is `package sharia` for the unit tests; the scenario test goes in the `sharia_test` package file alongside engine_test.go, e.g. `sharia/ijarah_scenario_test.go`)

- [ ] **Step 1: Write `sharia/ijarah_scenario_test.go`** (package `sharia_test`)

```go
package sharia_test

import (
	"encoding/json"
	"testing"

	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/sharia"
)

func ijarahReq(id string) sharia.CreateRequest {
	raw, _ := json.Marshal(sharia.IjarahParams{
		AssetCode: "VHCL1", Cost: sharia.Monetary{Asset: "DZD.2", Amount: 10000000},
		Rent: sharia.Monetary{Asset: "DZD.2", Amount: 500000},
		Client: "@client:anis", Supplier: "@supplier:toyota",
		BankTreasury: "@bank:treasury", Periods: 24,
		FirstDue: "2026-07-01T00:00:00Z", PeriodDays: 30,
	})
	return sharia.CreateRequest{Type: sharia.TypeIjarah, ID: id, Params: raw}
}

func TestIjarahNominalCycle(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "ijr_nominal"
		income0 := balance(t, l, "@bank:income:ijarah", "DZD.2")
		depr0 := balance(t, l, "@bank:expense:depreciation", "DZD.2")
		treas0 := balance(t, l, "@bank:treasury", "DZD.2")

		fund(t, l, "@bank:treasury", "DZD.2", 10000000)
		fund(t, l, "@client:anis", "DZD.2", 12000000)

		_, sched, err := e.Create(ijarahReq(id))
		if err != nil {
			t.Fatal(err)
		}
		if len(sched) != 24 || sched[0].Amount != 500000 || sched[23].DepreciationPart != 416682 {
			t.Fatalf("bad schedule")
		}

		// lease before acquire → SS-9
		_, err = e.Transition(id, "lease", sharia.TransitionInput{})
		se := shariaErr(t, err, sharia.ErrShariaViolation)
		if se.StandardRef != sharia.RefSS9 {
			t.Fatalf("expected SS-9, got %s", se.StandardRef)
		}

		mustTransition(t, e, id, "acquire")
		assertBal(t, l, "@supplier:toyota", "DZD.2", 10000000)
		assertBal(t, l, "@contracts:"+id+":asset", "VHCL1", 1)
		assertBal(t, l, "@contracts:"+id+":asset", "DZD.2", 10000000)

		mustTransition(t, e, id, "lease")
		assertBal(t, l, "@client:anis:in_use", "VHCL1", 1)

		var res sharia.TransitionResult
		var paidDepr int64
		for seq := 1; seq <= 24; seq++ {
			res, err = e.Transition(id, "pay_rent", sharia.TransitionInput{Seq: seq})
			if err != nil {
				t.Fatalf("pay_rent %d: %v", seq, err)
			}
			paidDepr += sched[seq-1].DepreciationPart
			assertBal(t, l, "@contracts:"+id+":asset", "DZD.2", 10000000-paidDepr)
		}
		if res.NewState != sharia.StateCompleted {
			t.Fatalf("expected COMPLETED, got %s", res.NewState)
		}

		assertBal(t, l, "@contracts:"+id+":asset", "DZD.2", 0)
		assertBal(t, l, "@contracts:"+id+":asset", "VHCL1", 0)
		assertBal(t, l, "@bank:inventory:returned", "VHCL1", 1)
		assertBal(t, l, "@bank:income:ijarah", "DZD.2", income0+12000000)
		assertBal(t, l, "@bank:expense:depreciation", "DZD.2", depr0+10000000)
		assertBal(t, l, "@bank:treasury", "DZD.2", treas0+10000000-10000000+12000000)
		assertBal(t, l, "@contracts:"+id+":counterpart", "DZD.2", 0)

		events, valid, err := e.VerifyAudit(id)
		if err != nil || !valid {
			t.Fatalf("audit must verify: %v", err)
		}
		if len(events) == 0 || events[0].Event != sharia.EventCreated {
			t.Fatal("expected created first")
		}
	})
}

func TestIjarahPenaltyToIncomeRejected(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "ijr_pen"
		fund(t, l, "@bank:treasury", "DZD.2", 10000000)
		mustCreateReq(t, e, ijarahReq(id))
		mustTransition(t, e, id, "acquire")
		mustTransition(t, e, id, "lease")
		_, err := e.Transition(id, "late_penalty", sharia.TransitionInput{
			Seq: 1, Amount: 20000, Destination: "@bank:income:ijarah",
		})
		se := shariaErr(t, err, sharia.ErrShariaViolation)
		if se.StandardRef != sharia.RefSS3 {
			t.Fatalf("expected SS-3, got %s", se.StandardRef)
		}
	})
}

func TestIjarahIdempotentRent(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "ijr_idem"
		fund(t, l, "@bank:treasury", "DZD.2", 10000000)
		fund(t, l, "@client:anis", "DZD.2", 2000000)
		mustCreateReq(t, e, ijarahReq(id))
		mustTransition(t, e, id, "acquire")
		mustTransition(t, e, id, "lease")
		if _, err := e.Transition(id, "pay_rent", sharia.TransitionInput{Seq: 1}); err != nil {
			t.Fatal(err)
		}
		_, err := e.Transition(id, "pay_rent", sharia.TransitionInput{Seq: 1})
		shariaErr(t, err, sharia.ErrDuplicate)
	})
}
```

- [ ] **Step 2: Add `mustCreateReq` helper** to `engine_test.go` (next to
`mustCreate`): `func mustCreateReq(t *testing.T, e *sharia.Engine, req sharia.CreateRequest)` that calls `e.Create(req)` and fails on error. (`mustCreate` currently builds a Murabaha request; add the request-taking variant so both contracts can reuse it.)

- [ ] **Step 3: Run** `go test ./sharia/ -count=1` → PASS
- [ ] **Step 4: Run full suite** `go test ./... -count=1` → all green
- [ ] **Step 5: Commit** `git add sharia/ && git commit -m "test(sharia): ijarah end-to-end nominal cycle, SS-3, idempotence"`

---

### Task 12: API + docs

**Files:** `api/actions/contract_controller_test.go` (add ijarah case), `sharia/README.md`

- [ ] **Step 1:** The contract controller already delegates to the engine, which
now routes by type — no controller code change. Add an httptest case creating a
`"type":"ijarah"` contract and asserting 201 + state PROMISE, to lock the
no-change-needed behavior.
- [ ] **Step 2: Run** `go test ./api/actions/ -count=1` → PASS
- [ ] **Step 3:** Extend `sharia/README.md`: Ijarah FSM, accounts, IJ-x → SS-9/FAS-32 → test mapping, and a "multi-contract engine: add a kind, register it" note.
- [ ] **Step 4: Commit** `git add api/ sharia/README.md && git commit -m "docs(sharia): ijarah API example + multi-contract README"`

---

## Self-review notes

- **Spec coverage:** Part 1 refactor (T1–T5) ✓; IjarahParams+validation (T6) ✓;
  schedule fixed-rent+depreciation (T7) ✓ persisted (T8) ✓; FSM+gate+postings
  (T9–T10) ✓; accounts incl. no-counterpart (T9 postings, T11 asserts =0) ✓;
  pay_rent 3 postings + P4 last + COMPLETED (T9/T10/T11) ✓; IJ-1 SS-9 (T11) ✓;
  IJ-5 SS-3 (T11) ✓; idempotence (T11) ✓; FAS-32-simplified label in audit
  payload (T10) ✓; multi-contract (engine generic, both kinds registered) ✓;
  README (T12) ✓.
- **Documented v1 limitation:** asset-return shares the last pay_rent transition
  (no standalone audit event) — noted in spec §2.5, acceptable for v1.
- **Type consistency:** `ContractKind` signature identical across kind.go,
  murabaha.go, ijarah.go (incl. `sched []Installment` param on Preconditions/
  BuildPlan). `Installment.DepreciationPart` added once (T7), persisted (T8),
  read in postings (T9). `RefFAS32` defined in T10 step 2.
- **Behavior preservation:** Part 1 verified by unchanged A–G after T4/T5.

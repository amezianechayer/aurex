# FaRl Guard v1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A declarative, hot-reloadable rule layer evaluated at the single chokepoint every state mutation crosses (`ledger.Commit`), that denies (abort, zero state change) or monitors (log, allow) transactions — each rule carrying its reason and citation.

**Architecture:** New `guard/` package (sibling of `sharia/`) with a `ContractKind`-style boundary: pure rule matchers + an `Engine` holding an in-memory rule snapshot under `RWMutex`. The engine's `Evaluate(view, txs, netFlows)` is called inside `ledger.Commit` right before `SaveTransactions`. Deny → write the deny `guard_event` via an autonomous store Exec, then return the error (SaveTransactions never runs → zero state change, proof survives). Monitor → events returned to the ledger and written best-effort after a successful commit. Per-ledger `guard_rules`/`guard_events` tables (migration v004). Default: no rules = current behavior strictly unchanged.

**Tech Stack:** Go 1.16, go-sqlbuilder, gin, fx. No new dependency.

**Spec:** `docs/superpowers/specs/2026-06-13-farl-guard-v1-design.md`
**Branch:** `feature/farl-guard-v1` (already created). NEVER commit to main.

**Import graph (no cycle):** `guard` → `core` only. `storage` → `guard` (interface) + `sharia` + `core`. `ledger` → `storage` + `guard` + `core`.

---

## File map

| File | Responsibility |
|---|---|
| `guard/types.go` (new) | `Rule`, `GuardEvent`, rule-kind + action constants, per-kind param structs |
| `guard/errors.go` (new) | typed `Error` (`ERR_GUARD_DENIED`/`ERR_INVALID_RULE`/`ERR_NOT_FOUND`) + HTTP mapping |
| `guard/rules.go` (new) | `scopeMatch`, the 3 kind matchers, `ValidateRule`, `evalRule` |
| `guard/store.go` (new) | `GuardStore` interface + `LedgerView` |
| `guard/engine.go` (new) | `Engine`: `Reload`, `Evaluate` (RWMutex snapshot, deny-event write, monitor collect) |
| `guard/*_test.go` (new) | matcher/validation/engine unit tests |
| `storage/sqlite/migration/v004.sql` (new) | `guard_rules` + `guard_events` (sqlite) |
| `storage/postgres/migration/v004.sql` (new) | same (postgres) |
| `storage/sqlite/guard.go` (new) | `GuardStore` impl (sqlite) |
| `storage/postgres/guard.go` (new) | `GuardStore` impl (postgres) |
| `storage/sqlite/guard_test.go` (new) | store round-trip |
| `storage/storage.go` (modify) | embed `guard.GuardStore` in `Store` |
| `ledger/ledger.go` (modify) | build/hold `guard.Engine`; hook in `Commit`; `Reload` at start |
| `api/actions/guard_controller.go` (new) | rule CRUD + events |
| `api/actions/controllers.go`, `api/routes/routes.go`, `api/api.go` (modify) | wire controller |

---

### Task 1: Guard types and errors

**Files:** Create `guard/types.go`, `guard/errors.go`

- [ ] **Step 1: Write `guard/types.go`**

```go
package guard

import "encoding/json"

// Rule kinds (v1)
const (
	KindAmountCap     = "amount_cap"
	KindAccountList   = "account_list"
	KindAssetRestrict = "asset_restrict"

	ActionDeny    = "deny"
	ActionMonitor = "monitor"
)

// Rule is a declarative guard rule, stored as JSON params per kind.
type Rule struct {
	ID          string          `json:"id"`
	Kind        string          `json:"kind"`
	Params      json.RawMessage `json:"params"`
	Action      string          `json:"action"`       // deny | monitor
	Reason      string          `json:"reason"`       // always required
	StandardRef string          `json:"standard_ref"` // required when action == deny
	Enabled     bool            `json:"enabled"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// GuardEvent records one rule firing (deny or monitor).
type GuardEvent struct {
	ID          int64  `json:"id"`
	RuleID      string `json:"rule_id"`
	Action      string `json:"action"`
	Reason      string `json:"reason"`
	StandardRef string `json:"standard_ref"`
	TxReference string `json:"tx_reference"`
	Payload     string `json:"payload"`
	CreatedAt   string `json:"created_at"`
}

// Per-kind param structs.
type AmountCapParams struct {
	Scope string `json:"scope"`
	Asset string `json:"asset"`
	Max   int64  `json:"max"`
	Basis string `json:"basis"` // "posting" | "net_outflow"
}

type AccountListParams struct {
	Mode     string   `json:"mode"`     // "block" | "allow"
	Side     string   `json:"side"`     // "source" | "destination" | "either"
	Patterns []string `json:"patterns"`
}

type AssetRestrictParams struct {
	Scope  string   `json:"scope"`
	Mode   string   `json:"mode"` // "only" | "never"
	Assets []string `json:"assets"`
}
```

- [ ] **Step 2: Write `guard/errors.go`**

```go
package guard

import "net/http"

const (
	ErrGuardDenied = "ERR_GUARD_DENIED"
	ErrInvalidRule = "ERR_INVALID_RULE"
	ErrNotFound    = "ERR_NOT_FOUND"
)

type Error struct {
	Code        string `json:"error"`
	Message     string `json:"message"`
	StandardRef string `json:"standard_ref,omitempty"`
	RuleID      string `json:"rule_id,omitempty"`
}

func (e *Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

func (e *Error) HTTPStatus() int {
	switch e.Code {
	case ErrInvalidRule:
		return http.StatusBadRequest
	case ErrNotFound:
		return http.StatusNotFound
	case ErrGuardDenied:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func newError(code, message string) *Error { return &Error{Code: code, Message: message} }
```

- [ ] **Step 3: Verify + commit** `go build ./guard/ 2>&1 || true` (package has no consumers yet — `go build ./...` still succeeds). Then:
```bash
git add guard/types.go guard/errors.go
git commit -m "feat(guard): rule and event types, typed errors"
```

---

### Task 2: Scope matching and the three rule matchers

**Files:** Create `guard/rules.go`, `guard/rules_test.go`

- [ ] **Step 1: Write failing tests** `guard/rules_test.go`

```go
package guard

import (
	"encoding/json"
	"testing"

	"github.com/amezianechayer/corren/core"
)

func TestScopeMatch(t *testing.T) {
	cases := []struct {
		pattern, account string
		want             bool
	}{
		{"@client:*", "@client:anis", true},
		{"@client:*", "@clientx", false},
		{"@contracts:*", "@contracts:mur1:receivable", true},
		{"@bank:treasury", "@bank:treasury", true},
		{"@bank:treasury", "@bank:income", false},
		{"*", "@anything", true},
	}
	for _, c := range cases {
		if got := scopeMatch(c.pattern, c.account); got != c.want {
			t.Errorf("scopeMatch(%q,%q)=%v want %v", c.pattern, c.account, got, c.want)
		}
	}
}

func tx(postings ...core.Posting) []core.Transaction {
	return []core.Transaction{{Postings: postings}}
}

func netFlows(txs []core.Transaction) map[string]map[string]int64 {
	rf := map[string]map[string]int64{}
	for _, t := range txs {
		for _, p := range t.Postings {
			if rf[p.Source] == nil {
				rf[p.Source] = map[string]int64{}
			}
			rf[p.Source][p.Asset] += p.Amount
			if rf[p.Destination] == nil {
				rf[p.Destination] = map[string]int64{}
			}
			rf[p.Destination][p.Asset] -= p.Amount
		}
	}
	return rf
}

func TestAmountCapPostingBasis(t *testing.T) {
	raw, _ := json.Marshal(AmountCapParams{Scope: "@client:*", Asset: "DZD.2", Max: 1000, Basis: "posting"})
	r := Rule{Kind: KindAmountCap, Params: raw}
	txs := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 1500})
	if matched, _ := evalRule(r, txs, netFlows(txs)); !matched {
		t.Fatal("posting 1500 > cap 1000 from @client:* must match")
	}
	txs2 := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 800})
	if matched, _ := evalRule(r, txs2, netFlows(txs2)); matched {
		t.Fatal("posting 800 <= cap 1000 must not match")
	}
	// wrong asset is ignored
	txs3 := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "EUR.2", Amount: 5000})
	if matched, _ := evalRule(r, txs3, netFlows(txs3)); matched {
		t.Fatal("different asset must not match")
	}
}

func TestAmountCapNetOutflow(t *testing.T) {
	raw, _ := json.Marshal(AmountCapParams{Scope: "@client:anis", Asset: "DZD.2", Max: 1000, Basis: "net_outflow"})
	r := Rule{Kind: KindAmountCap, Params: raw}
	// sends 1500, receives 700 → net outflow 800 ≤ 1000 → no match
	txs := tx(
		core.Posting{Source: "@client:anis", Destination: "@a", Asset: "DZD.2", Amount: 1500},
		core.Posting{Source: "@b", Destination: "@client:anis", Asset: "DZD.2", Amount: 700},
	)
	if matched, _ := evalRule(r, txs, netFlows(txs)); matched {
		t.Fatal("net outflow 800 <= 1000 must not match")
	}
	// sends 1500, receives 0 → net 1500 > 1000 → match
	txs2 := tx(core.Posting{Source: "@client:anis", Destination: "@a", Asset: "DZD.2", Amount: 1500})
	if matched, _ := evalRule(r, txs2, netFlows(txs2)); !matched {
		t.Fatal("net outflow 1500 > 1000 must match")
	}
}

func TestAccountListBlockAndAllow(t *testing.T) {
	block, _ := json.Marshal(AccountListParams{Mode: "block", Side: "destination", Patterns: []string{"@sanctioned:*"}})
	rb := Rule{Kind: KindAccountList, Params: block}
	txs := tx(core.Posting{Source: "@client:anis", Destination: "@sanctioned:x", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(rb, txs, netFlows(txs)); !matched {
		t.Fatal("blocklisted destination must match")
	}
	ok := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(rb, ok, netFlows(ok)); matched {
		t.Fatal("non-blocklisted destination must not match")
	}

	// allow mode: matches (i.e. should be denied) when a posting side is NOT in the allowlist
	allow, _ := json.Marshal(AccountListParams{Mode: "allow", Side: "destination", Patterns: []string{"@bank:*"}})
	ra := Rule{Kind: KindAccountList, Params: allow}
	bad := tx(core.Posting{Source: "@client:anis", Destination: "@stranger:y", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(ra, bad, netFlows(bad)); !matched {
		t.Fatal("destination outside allowlist must match (deny)")
	}
	good := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(ra, good, netFlows(good)); matched {
		t.Fatal("destination in allowlist must not match")
	}
}

func TestAssetRestrict(t *testing.T) {
	only, _ := json.Marshal(AssetRestrictParams{Scope: "@client:*", Mode: "only", Assets: []string{"DZD.2"}})
	r := Rule{Kind: KindAssetRestrict, Params: only}
	bad := tx(core.Posting{Source: "@client:anis", Destination: "@a", Asset: "EUR.2", Amount: 10})
	if matched, _ := evalRule(r, bad, netFlows(bad)); !matched {
		t.Fatal("client transacting a non-allowed asset must match")
	}
	good := tx(core.Posting{Source: "@client:anis", Destination: "@a", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(r, good, netFlows(good)); matched {
		t.Fatal("client transacting the only-allowed asset must not match")
	}

	never, _ := json.Marshal(AssetRestrictParams{Scope: "@client:*", Mode: "never", Assets: []string{"GOLD"}})
	rn := Rule{Kind: KindAssetRestrict, Params: never}
	g := tx(core.Posting{Source: "@client:anis", Destination: "@a", Asset: "GOLD", Amount: 1})
	if matched, _ := evalRule(rn, g, netFlows(g)); !matched {
		t.Fatal("never-allowed asset must match")
	}
}
```

- [ ] **Step 2: Run** `go test ./guard/ -run 'TestScopeMatch|TestAmountCap|TestAccountList|TestAssetRestrict'` → FAIL (undefined scopeMatch/evalRule)

- [ ] **Step 3: Implement `guard/rules.go`**

```go
package guard

import (
	"encoding/json"
	"strings"

	"github.com/amezianechayer/corren/core"
)

// scopeMatch: prefix glob. "@client:*" matches "@client:anis"; "*" matches all;
// otherwise exact match. No regex (lexer-safe, YAGNI).
func scopeMatch(pattern, account string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(account, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == account
}

func anyMatch(patterns []string, account string) bool {
	for _, p := range patterns {
		if scopeMatch(p, account) {
			return true
		}
	}
	return false
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// evalRule reports whether the rule MATCHES (i.e. should fire) against the
// proposed transactions. A match means: deny (if action=deny) or record a
// monitor event (if action=monitor). The string is a short detail for the
// event payload. Pure: reads only txs + netFlows.
func evalRule(r Rule, txs []core.Transaction, nf map[string]map[string]int64) (bool, string) {
	switch r.Kind {
	case KindAmountCap:
		var p AmountCapParams
		if json.Unmarshal(r.Params, &p) != nil {
			return false, ""
		}
		if p.Basis == "net_outflow" {
			for acct, assets := range nf {
				if !scopeMatch(p.Scope, acct) {
					continue
				}
				if out := assets[p.Asset]; out > p.Max {
					return true, "net_outflow " + acct
				}
			}
			return false, ""
		}
		// default: posting basis
		for _, t := range txs {
			for _, post := range t.Postings {
				if post.Asset == p.Asset && scopeMatch(p.Scope, post.Source) && post.Amount > p.Max {
					return true, "posting " + post.Source
				}
			}
		}
		return false, ""

	case KindAccountList:
		var p AccountListParams
		if json.Unmarshal(r.Params, &p) != nil {
			return false, ""
		}
		sideHit := func(post core.Posting) bool {
			switch p.Side {
			case "source":
				return anyMatch(p.Patterns, post.Source)
			case "destination":
				return anyMatch(p.Patterns, post.Destination)
			default: // either
				return anyMatch(p.Patterns, post.Source) || anyMatch(p.Patterns, post.Destination)
			}
		}
		for _, t := range txs {
			for _, post := range t.Postings {
				hit := sideHit(post)
				if p.Mode == "block" && hit {
					return true, "blocked " + post.Source + "->" + post.Destination
				}
				if p.Mode == "allow" && !hit {
					return true, "not allowlisted " + post.Source + "->" + post.Destination
				}
			}
		}
		return false, ""

	case KindAssetRestrict:
		var p AssetRestrictParams
		if json.Unmarshal(r.Params, &p) != nil {
			return false, ""
		}
		for _, t := range txs {
			for _, post := range t.Postings {
				inScope := scopeMatch(p.Scope, post.Source) || scopeMatch(p.Scope, post.Destination)
				if !inScope {
					continue
				}
				listed := contains(p.Assets, post.Asset)
				if p.Mode == "only" && !listed {
					return true, "asset " + post.Asset + " not allowed"
				}
				if p.Mode == "never" && listed {
					return true, "asset " + post.Asset + " forbidden"
				}
			}
		}
		return false, ""
	}
	return false, ""
}
```

- [ ] **Step 4: Run** `go test ./guard/` → PASS
- [ ] **Step 5: Commit** `git add guard/rules.go guard/rules_test.go && git commit -m "feat(guard): scope glob + amount_cap/account_list/asset_restrict matchers"`

---

### Task 3: Rule validation

**Files:** Modify `guard/rules.go`, `guard/rules_test.go`

- [ ] **Step 1: Add failing tests**

```go
func TestValidateRule(t *testing.T) {
	ok := Rule{Kind: KindAmountCap, Action: ActionDeny, Reason: "limit", StandardRef: "POLICY-1",
		Params: mustJSON(AmountCapParams{Scope: "@client:*", Asset: "DZD.2", Max: 1000, Basis: "posting"})}
	if err := ValidateRule(&ok); err != nil {
		t.Fatalf("valid rule rejected: %v", err)
	}

	bad := []Rule{
		{Kind: "bogus", Action: ActionDeny, Reason: "x", StandardRef: "P"},                          // unknown kind
		{Kind: KindAmountCap, Action: ActionDeny, Reason: "x", Params: mustJSON(AmountCapParams{Scope: "@c:*", Asset: "DZD.2", Max: 1, Basis: "posting"})}, // deny without standard_ref
		{Kind: KindAmountCap, Action: ActionMonitor, StandardRef: "P", Params: mustJSON(AmountCapParams{Scope: "@c:*", Asset: "DZD.2", Max: 1, Basis: "posting"})},  // no reason
		{Kind: KindAmountCap, Action: "wat", Reason: "x", StandardRef: "P", Params: mustJSON(AmountCapParams{Scope: "@c:*", Asset: "DZD.2", Max: 1, Basis: "posting"})}, // bad action
		{Kind: KindAmountCap, Action: ActionDeny, Reason: "x", StandardRef: "P", Params: []byte(`{"basis":"bogus"}`)}, // bad basis
	}
	for i, r := range bad {
		r := r
		if err := ValidateRule(&r); err == nil {
			t.Fatalf("bad rule %d accepted", i)
		}
	}
}

func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
```

- [ ] **Step 2: Run** → FAIL (undefined ValidateRule)
- [ ] **Step 3: Implement `ValidateRule` in `guard/rules.go`**

```go
// ValidateRule checks a rule is well-formed before it is stored. Critically:
// a deny rule MUST carry a standard_ref (an unexplained refusal is an audit
// hole), and reason is always required.
func ValidateRule(r *Rule) error {
	if r.Action != ActionDeny && r.Action != ActionMonitor {
		return newError(ErrInvalidRule, "action must be deny or monitor")
	}
	if strings.TrimSpace(r.Reason) == "" {
		return newError(ErrInvalidRule, "reason is required")
	}
	if r.Action == ActionDeny && strings.TrimSpace(r.StandardRef) == "" {
		return newError(ErrInvalidRule, "standard_ref is required for a deny rule")
	}
	switch r.Kind {
	case KindAmountCap:
		var p AmountCapParams
		if err := json.Unmarshal(r.Params, &p); err != nil {
			return newError(ErrInvalidRule, "invalid amount_cap params")
		}
		if p.Basis != "posting" && p.Basis != "net_outflow" {
			return newError(ErrInvalidRule, "amount_cap basis must be posting or net_outflow")
		}
		if p.Scope == "" || p.Asset == "" || p.Max <= 0 {
			return newError(ErrInvalidRule, "amount_cap requires scope, asset and max > 0")
		}
	case KindAccountList:
		var p AccountListParams
		if err := json.Unmarshal(r.Params, &p); err != nil {
			return newError(ErrInvalidRule, "invalid account_list params")
		}
		if p.Mode != "block" && p.Mode != "allow" {
			return newError(ErrInvalidRule, "account_list mode must be block or allow")
		}
		if len(p.Patterns) == 0 {
			return newError(ErrInvalidRule, "account_list requires patterns")
		}
	case KindAssetRestrict:
		var p AssetRestrictParams
		if err := json.Unmarshal(r.Params, &p); err != nil {
			return newError(ErrInvalidRule, "invalid asset_restrict params")
		}
		if p.Mode != "only" && p.Mode != "never" {
			return newError(ErrInvalidRule, "asset_restrict mode must be only or never")
		}
		if p.Scope == "" || len(p.Assets) == 0 {
			return newError(ErrInvalidRule, "asset_restrict requires scope and assets")
		}
	default:
		return newError(ErrInvalidRule, "unknown rule kind "+r.Kind)
	}
	return nil
}
```

- [ ] **Step 4: Run** `go test ./guard/` → PASS
- [ ] **Step 5: Commit** `git add guard/rules.go guard/rules_test.go && git commit -m "feat(guard): rule validation (deny requires standard_ref)"`

---

### Task 4: GuardStore interface + migrations + both flavors

**Files:** Create `guard/store.go`, `storage/sqlite/migration/v004.sql`, `storage/postgres/migration/v004.sql`, `storage/sqlite/guard.go`, `storage/postgres/guard.go`, `storage/sqlite/guard_test.go`

- [ ] **Step 1: Write `guard/store.go`**

```go
package guard

import "github.com/amezianechayer/corren/core"

// LedgerView is the minimal read surface a rule MIGHT need, passed to Evaluate
// to keep the signature stable and avoid the guard→ledger import cycle. The v1
// rules don't use it; v2 rules (e.g. balance-dependent) can.
type LedgerView interface {
	GetAccount(address string) (core.Account, error)
}

// GuardStore persists rules and events. Implemented by storage/sqlite and
// storage/postgres, embedded into storage.Store like sharia.ShariaStore.
type GuardStore interface {
	SaveRule(r Rule) error
	UpdateRule(r Rule) error
	DeleteRule(id string) error
	GetRule(id string) (Rule, error)
	ListRules() ([]Rule, error)

	AppendGuardEvent(e GuardEvent) error
	ListGuardEvents(limit, offset int) ([]GuardEvent, error)
}
```

- [ ] **Step 2: Migrations.** `storage/sqlite/migration/v004.sql`:

```sql
--statement
CREATE TABLE IF NOT EXISTS guard_rules (
  "id"           varchar,
  "kind"         varchar,
  "params"       varchar,
  "action"       varchar,
  "reason"       varchar,
  "standard_ref" varchar,
  "enabled"      integer,
  "created_at"   varchar,
  "updated_at"   varchar,
  UNIQUE("id")
);
--statement
CREATE TABLE IF NOT EXISTS guard_events (
  "id"           integer,
  "rule_id"      varchar,
  "action"       varchar,
  "reason"       varchar,
  "standard_ref" varchar,
  "tx_reference" varchar,
  "payload"      varchar,
  "created_at"   varchar
);
--statement
CREATE INDEX IF NOT EXISTS 'guard_events_created' ON "guard_events" ("created_at");
```

`storage/postgres/migration/v004.sql` (schema-qualified, `integer`→`bigint`):

```sql
--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".guard_rules (
  "id"           varchar,
  "kind"         varchar,
  "params"       varchar,
  "action"       varchar,
  "reason"       varchar,
  "standard_ref" varchar,
  "enabled"      integer,
  "created_at"   varchar,
  "updated_at"   varchar,
  UNIQUE("id")
);
--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".guard_events (
  "id"           bigint,
  "rule_id"      varchar,
  "action"       varchar,
  "reason"       varchar,
  "standard_ref" varchar,
  "tx_reference" varchar,
  "payload"      varchar,
  "created_at"   varchar
);
--statement
CREATE INDEX IF NOT EXISTS guard_events_created ON "VAR_LEDGER_NAME".guard_events ("created_at");
```

- [ ] **Step 3: Failing store test** `storage/sqlite/guard_test.go` (mirror `sharia_test.go` setup: it shares `package sqlite`, reuses `NewStore`/`Initialize`; use a unique db_name to avoid clashing — note the existing `TestMain` sets `storage.sqlite.db_name = "sharia_store"`, so guard tests run in the same sqlite test binary and DB; use distinct ledger name `gtest`).

```go
package sqlite

import (
	"encoding/json"
	"testing"

	"github.com/amezianechayer/corren/guard"
)

func withGuardStore(t *testing.T, f func(s *SQLiteStore)) {
	t.Helper()
	s, err := NewStore("gtest")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	f(s)
}

func TestGuardRuleRoundTrip(t *testing.T) {
	withGuardStore(t, func(s *SQLiteStore) {
		params, _ := json.Marshal(guard.AmountCapParams{Scope: "@client:*", Asset: "DZD.2", Max: 1000, Basis: "posting"})
		r := guard.Rule{ID: "rule-1", Kind: guard.KindAmountCap, Params: params,
			Action: guard.ActionDeny, Reason: "limit", StandardRef: "POLICY-1", Enabled: true,
			CreatedAt: "2026-06-13T00:00:00Z", UpdatedAt: "2026-06-13T00:00:00Z"}
		if err := s.SaveRule(r); err != nil {
			t.Fatal(err)
		}
		got, err := s.GetRule("rule-1")
		if err != nil || got.Kind != guard.KindAmountCap || !got.Enabled || got.StandardRef != "POLICY-1" {
			t.Fatalf("got %+v err %v", got, err)
		}
		r.Enabled = false
		if err := s.UpdateRule(r); err != nil {
			t.Fatal(err)
		}
		got, _ = s.GetRule("rule-1")
		if got.Enabled {
			t.Fatal("expected disabled after update")
		}
		list, err := s.ListRules()
		if err != nil || len(list) == 0 {
			t.Fatalf("list: %v %v", list, err)
		}
		if err := s.DeleteRule("rule-1"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.GetRule("rule-1"); err == nil {
			t.Fatal("expected not found after delete")
		}
	})
}

func TestGuardEventRoundTrip(t *testing.T) {
	withGuardStore(t, func(s *SQLiteStore) {
		e := guard.GuardEvent{RuleID: "rule-1", Action: guard.ActionDeny, Reason: "limit",
			StandardRef: "POLICY-1", TxReference: "tx:1", Payload: `{"x":1}`, CreatedAt: "2026-06-13T00:00:00Z"}
		if err := s.AppendGuardEvent(e); err != nil {
			t.Fatal(err)
		}
		events, err := s.ListGuardEvents(10, 0)
		if err != nil || len(events) == 0 {
			t.Fatalf("events: %v %v", events, err)
		}
		if events[len(events)-1].RuleID != "rule-1" {
			t.Fatalf("unexpected event: %+v", events[0])
		}
	})
}
```

- [ ] **Step 4: Run** `go test ./storage/sqlite/ -run TestGuard` → FAIL (SaveRule undefined)
- [ ] **Step 5: Implement `storage/sqlite/guard.go`** — follow the exact style of `storage/sqlite/sharia.go` (go-sqlbuilder, `sqlbuilder.SQLite`, `s.db`). `SaveRule`/`UpdateRule`/`GetRule`/`ListRules`/`DeleteRule` map booleans to int (`enabled`: `1`/`0`); `GetRule` maps `sql.ErrNoRows`→ a `guard` not-found error (`&guard.Error{Code: guard.ErrNotFound}`). `AppendGuardEvent` assigns `id = count+1` via `SELECT COUNT(*) FROM guard_events` (same pattern as the sharia audit). `ListGuardEvents` orders by `id ASC` with limit/offset. Scan `enabled` into an int then set `r.Enabled = (n == 1)`.

- [ ] **Step 6: Implement `storage/postgres/guard.go`** — same methods, `sqlbuilder.PostgreSQL`, `s.table("guard_rules")`, `s.Conn().Exec/QueryRow/Query(context.Background(), ...)`, mirror `storage/postgres/sharia.go`. Map `pgx.ErrNoRows` → not-found.

- [ ] **Step 7: Run** `go test ./storage/sqlite/ -run TestGuard` → PASS. `go build ./...`.
- [ ] **Step 8: Commit** `git add guard/store.go storage/ && git commit -m "feat(storage): GuardStore (sqlite+postgres) and migration v004"`

---

### Task 5: The Guard engine

**Files:** Create `guard/engine.go`, `guard/engine_test.go`

- [ ] **Step 1: Failing tests** `guard/engine_test.go` (uses an in-memory fake store)

```go
package guard

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/amezianechayer/corren/core"
)

type fakeStore struct {
	mu     sync.Mutex
	rules  []Rule
	events []GuardEvent
}

func (f *fakeStore) SaveRule(r Rule) error   { f.rules = append(f.rules, r); return nil }
func (f *fakeStore) UpdateRule(r Rule) error { return nil }
func (f *fakeStore) DeleteRule(id string) error { return nil }
func (f *fakeStore) GetRule(id string) (Rule, error) { return Rule{}, &Error{Code: ErrNotFound} }
func (f *fakeStore) ListRules() ([]Rule, error) { return f.rules, nil }
func (f *fakeStore) AppendGuardEvent(e GuardEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, e)
	return nil
}
func (f *fakeStore) ListGuardEvents(limit, offset int) ([]GuardEvent, error) { return f.events, nil }

func capRule(id, action string, max int64) Rule {
	p, _ := json.Marshal(AmountCapParams{Scope: "@client:*", Asset: "DZD.2", Max: max, Basis: "posting"})
	return Rule{ID: id, Kind: KindAmountCap, Params: p, Action: action,
		Reason: "limit", StandardRef: "POLICY-1", Enabled: true}
}

func over(amount int64) []core.Transaction {
	return []core.Transaction{{Reference: "tx:over", Postings: []core.Posting{
		{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: amount}}}}
}

func TestEngineDenyWritesEventAndErrors(t *testing.T) {
	fs := &fakeStore{}
	fs.rules = []Rule{capRule("r1", ActionDeny, 1000)}
	e := NewEngine(fs)
	if err := e.Reload(); err != nil {
		t.Fatal(err)
	}
	mon, err := e.Evaluate(nil, over(1500), nil)
	if err == nil {
		t.Fatal("expected deny error")
	}
	ge, ok := err.(*Error)
	if !ok || ge.Code != ErrGuardDenied || ge.StandardRef != "POLICY-1" || ge.RuleID != "r1" {
		t.Fatalf("bad deny error: %+v", err)
	}
	if len(mon) != 0 {
		t.Fatal("deny must not return monitor events")
	}
	// the deny event was written immediately (proof survives a later rollback)
	if len(fs.events) != 1 || fs.events[0].Action != ActionDeny {
		t.Fatalf("deny event not written: %+v", fs.events)
	}
}

func TestEngineMonitorReturnsEventAllows(t *testing.T) {
	fs := &fakeStore{}
	fs.rules = []Rule{capRule("r1", ActionMonitor, 1000)}
	e := NewEngine(fs)
	e.Reload()
	mon, err := e.Evaluate(nil, over(1500), nil)
	if err != nil {
		t.Fatalf("monitor must not deny: %v", err)
	}
	if len(mon) != 1 || mon[0].Action != ActionMonitor || mon[0].RuleID != "r1" {
		t.Fatalf("expected one monitor event, got %+v", mon)
	}
	// monitor events are NOT written by Evaluate (the ledger writes them post-commit)
	if len(fs.events) != 0 {
		t.Fatal("monitor events must not be written by Evaluate")
	}
}

func TestEngineDisabledAndNoMatch(t *testing.T) {
	fs := &fakeStore{}
	r := capRule("r1", ActionDeny, 1000)
	r.Enabled = false
	fs.rules = []Rule{r}
	e := NewEngine(fs)
	e.Reload()
	if _, err := e.Evaluate(nil, over(1500), nil); err != nil {
		t.Fatal("disabled rule must not fire")
	}
	// no rules at all → no-op
	e2 := NewEngine(&fakeStore{})
	e2.Reload()
	if mon, err := e2.Evaluate(nil, over(99999), nil); err != nil || mon != nil {
		t.Fatalf("empty ruleset must be a no-op, got mon=%v err=%v", mon, err)
	}
}

func TestEngineReloadRaceSafe(t *testing.T) {
	fs := &fakeStore{}
	fs.rules = []Rule{capRule("r1", ActionMonitor, 1000)}
	e := NewEngine(fs)
	e.Reload()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); e.Evaluate(nil, over(1500), nil) }()
		go func() { defer wg.Done(); e.Reload() }()
	}
	wg.Wait()
}
```

- [ ] **Step 2: Run** `go test -race ./guard/ -run TestEngine` → FAIL (NewEngine undefined)
- [ ] **Step 3: Implement `guard/engine.go`**

```go
package guard

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/amezianechayer/corren/core"
)

type Engine struct {
	store GuardStore
	mu    sync.RWMutex
	rules []Rule
}

func NewEngine(store GuardStore) *Engine { return &Engine{store: store} }

func now() string { return time.Now().UTC().Format(time.RFC3339) }

// Reload refreshes the in-memory rule snapshot from the store. Called at
// startup and after every admin mutation (hot-reload).
func (e *Engine) Reload() error {
	rules, err := e.store.ListRules()
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.rules = rules
	e.mu.Unlock()
	return nil
}

func firstReference(txs []core.Transaction) string {
	for _, t := range txs {
		if t.Reference != "" {
			return t.Reference
		}
	}
	return ""
}

// Evaluate is the hook ledger.Commit calls just before SaveTransactions.
// On a matching deny rule it writes the deny guard_event immediately (its own
// store Exec — survives the rollback since SaveTransactions never runs) and
// returns *Error. Matching monitor rules are accumulated and returned for the
// ledger to persist AFTER a successful commit. No I/O unless a rule fires.
func (e *Engine) Evaluate(view LedgerView, txs []core.Transaction,
	netFlows map[string]map[string]int64) ([]GuardEvent, error) {

	e.mu.RLock()
	snapshot := e.rules
	e.mu.RUnlock()

	if len(snapshot) == 0 {
		return nil, nil
	}

	ref := firstReference(txs)
	var monitorEvents []GuardEvent

	for _, r := range snapshot {
		if !r.Enabled {
			continue
		}
		matched, detail := evalRule(r, txs, netFlows)
		if !matched {
			continue
		}
		payload, _ := json.Marshal(map[string]string{"detail": detail})
		ev := GuardEvent{
			RuleID: r.ID, Action: r.Action, Reason: r.Reason,
			StandardRef: r.StandardRef, TxReference: ref, Payload: string(payload),
			CreatedAt: now(),
		}
		if r.Action == ActionDeny {
			// proof must survive the (non-)commit: write now, then abort.
			if werr := e.store.AppendGuardEvent(ev); werr != nil {
				log.Printf("guard: failed to persist deny event for rule %s: %v", r.ID, werr)
			}
			return nil, &Error{Code: ErrGuardDenied, Message: r.Reason,
				StandardRef: r.StandardRef, RuleID: r.ID}
		}
		monitorEvents = append(monitorEvents, ev)
	}
	return monitorEvents, nil
}

// WriteMonitorEvents persists monitor events after a successful commit
// (best-effort, same durability model as the sharia audit trail).
func (e *Engine) WriteMonitorEvents(events []GuardEvent) {
	for _, ev := range events {
		if err := e.store.AppendGuardEvent(ev); err != nil {
			log.Printf("guard: failed to persist monitor event for rule %s: %v", ev.RuleID, err)
		}
	}
}
```

- [ ] **Step 4: Run** `go test -race ./guard/` → PASS
- [ ] **Step 5: Commit** `git add guard/engine.go guard/engine_test.go && git commit -m "feat(guard): engine with RWMutex snapshot, deny-event durability, monitor collect"`

---

### Task 6: Wire Guard into storage.Store and ledger.Commit

**Files:** Modify `storage/storage.go`, `ledger/ledger.go`, `ledger/virtual_test.go` (or a new `ledger/guard_test.go`)

- [ ] **Step 1: Embed `guard.GuardStore` in `storage.Store`** — in `storage/storage.go`, add the import and embed:

```go
import (
	// ...existing...
	"github.com/amezianechayer/corren/guard"
)

type Store interface {
	// ...existing methods...
	sharia.ShariaStore
	guard.GuardStore
}
```

`go build ./...` will now fail until both flavors satisfy it — they do (Task 4 added the methods). Confirm build passes.

- [ ] **Step 2: Add the guard engine to `Ledger` and hook `Commit`.** In `ledger/ledger.go`:
  - import `"github.com/amezianechayer/corren/guard"`.
  - add a field `guard *guard.Engine` to the `Ledger` struct.
  - in `NewLedger`, after the store is initialized, build and load it:

```go
	l := &Ledger{
		store:       store,
		name:        name,
		_lastMetaID: -1,
	}
	l.guard = guard.NewEngine(store)
	if err := l.guard.Reload(); err != nil {
		log.Printf("guard: initial reload failed for ledger %s: %v", name, err)
	}
```

  - in `Commit`, replace the tail (the `err := l.store.SaveTransactions(ts)` ... `return ts, err` part, AFTER the `for addr := range rf` balance loop) with:

```go
	monitorEvents, gerr := l.guard.Evaluate(l, ts, rf)
	if gerr != nil {
		return ts, gerr // deny: SaveTransactions never runs → zero state change
	}

	err := l.store.SaveTransactions(ts)
	if err == nil && len(monitorEvents) > 0 {
		l.guard.WriteMonitorEvents(monitorEvents)
	}
	l._last = &ts[len(ts)-1]
	return ts, err
```

  - add a `Guard()` accessor: `func (l *Ledger) Guard() *guard.Engine { return l.guard }` (the API controller uses it to Reload after mutations). `*ledger.Ledger` already has `GetAccount`, so it satisfies `guard.LedgerView` — passing `l` as the view compiles.

- [ ] **Step 3: Add the keystone integration test** `ledger/guard_test.go` (package `ledger`, reuses the `with` helper + `assertBalance` from existing ledger tests):

```go
package ledger

import (
	"encoding/json"
	"testing"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/guard"
)

func saveRuleAndReload(t *testing.T, l *Ledger, r guard.Rule) {
	t.Helper()
	if err := l.store.SaveRule(r); err != nil {
		t.Fatal(err)
	}
	if err := l.Guard().Reload(); err != nil {
		t.Fatal(err)
	}
}

func capRule(id, action string, max int64) guard.Rule {
	p, _ := json.Marshal(guard.AmountCapParams{Scope: "@client:*", Asset: "DZD.2", Max: max, Basis: "posting"})
	return guard.Rule{ID: id, Kind: guard.KindAmountCap, Params: p, Action: action,
		Reason: "over limit", StandardRef: "POLICY-1", Enabled: true,
		CreatedAt: "2026-06-13T00:00:00Z", UpdatedAt: "2026-06-13T00:00:00Z"}
}

func fundClient(t *testing.T, l *Ledger, amount int64) {
	t.Helper()
	_, err := l.Commit([]core.Transaction{{Postings: []core.Posting{
		{Source: core.WORLD, Destination: "@client:anis", Asset: "DZD.2", Amount: amount}}}})
	if err != nil {
		t.Fatal(err)
	}
}

// Deny: the tx is rejected, the ledger is unchanged, AND the deny guard_event
// is persisted despite SaveTransactions never running (proof survives rollback).
func TestGuardDenySurvivesRollback(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()
		fundClient(t, l, 5000)
		saveRuleAndReload(t, l, capRule("g-deny", guard.ActionDeny, 1000))

		_, err := l.Commit([]core.Transaction{{Reference: "pay:big", Postings: []core.Posting{
			{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 1500}}}})
		if err == nil {
			t.Fatal("expected guard deny")
		}
		ge, ok := err.(*guard.Error)
		if !ok || ge.Code != guard.ErrGuardDenied || ge.StandardRef != "POLICY-1" {
			t.Fatalf("bad deny error: %v", err)
		}
		// ledger unchanged: the 1500 never left the client
		assertBalance(t, l, "@client:anis", "DZD.2", 5000)
		assertBalance(t, l, "@bank:treasury", "DZD.2", 0)
		// proof persisted despite the abort
		events, err := l.store.ListGuardEvents(10, 0)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range events {
			if e.RuleID == "g-deny" && e.Action == guard.ActionDeny {
				found = true
			}
		}
		if !found {
			t.Fatal("deny guard_event must persist even though the tx rolled back")
		}
	})
}

// Monitor: the tx passes, and a monitor event is recorded.
func TestGuardMonitorAllows(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()
		fundClient(t, l, 5000)
		saveRuleAndReload(t, l, capRule("g-mon", guard.ActionMonitor, 1000))

		_, err := l.Commit([]core.Transaction{{Reference: "pay:watched", Postings: []core.Posting{
			{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 1500}}}})
		if err != nil {
			t.Fatalf("monitor must allow: %v", err)
		}
		assertBalance(t, l, "@bank:treasury", "DZD.2", 1500)
		events, _ := l.store.ListGuardEvents(10, 0)
		found := false
		for _, e := range events {
			if e.RuleID == "g-mon" && e.Action == guard.ActionMonitor {
				found = true
			}
		}
		if !found {
			t.Fatal("monitor guard_event must be recorded")
		}
	})
}

// Zero rules = strict no-op (a normal transfer still commits).
func TestGuardNoRulesNoOp(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()
		_, err := l.Commit([]core.Transaction{{Postings: []core.Posting{
			{Source: core.WORLD, Destination: "@x", Asset: "DZD.2", Amount: 42}}}})
		if err != nil {
			t.Fatal(err)
		}
		assertBalance(t, l, "@x", "DZD.2", 42)
	})
}
```

- [ ] **Step 4: Run** `go test ./ledger/ -run TestGuard -count=1 -v` → PASS, then `go test ./... -count=1` → all green (zero-rule default keeps every existing test unchanged).
- [ ] **Step 5: Commit** `git add storage/storage.go ledger/ledger.go ledger/guard_test.go && git commit -m "feat(ledger): evaluate guard rules at Commit (deny aborts, proof survives rollback)"`

---

### Task 7: Contract-engine coverage test (the decisive proof)

**Files:** Create `sharia/guard_coverage_test.go` (package `sharia_test`)

This proves a `@contracts:*`-scoped rule blocks a posting generated by the
contracts engine — i.e. Guard covers engine mutations, not just scripts/API.

- [ ] **Step 1: Write the test** (reuses `withEngine`/`fund`/`mustCreate` helpers from `engine_test.go`, and the murabaha `params` helper)

```go
package sharia_test

import (
	"encoding/json"
	"testing"

	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/sharia"
)

func TestGuardCoversContractEnginePostings(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "mur_guard"
		fund(t, l, "@bank:treasury", "SAR2", 20000000)

		// a deny rule scoped to contract receivable accounts, capping any
		// posting above 1,000,000 — the sell will create an 11,000,000 receivable
		p, _ := json.Marshal(guard.AmountCapParams{
			Scope: "@contracts:*", Asset: "SAR2", Max: 1000000, Basis: "posting"})
		rule := guard.Rule{ID: "g-contracts", Kind: guard.KindAmountCap, Params: p,
			Action: guard.ActionDeny, Reason: "contract posting over cap",
			StandardRef: "POLICY-CONTRACT", Enabled: true,
			CreatedAt: "2026-06-13T00:00:00Z", UpdatedAt: "2026-06-13T00:00:00Z"}
		if err := l.Store().SaveRule(rule); err != nil {
			t.Fatal(err)
		}
		if err := l.Guard().Reload(); err != nil {
			t.Fatal(err)
		}

		mustCreate(t, e, id, params("@client:ameziane", "@supplier:toyota", "@bank:treasury", 24))
		mustTransition(t, e, id, "acquire")

		// sell builds an 11,000,000 receivable posting → guard must deny it
		_, err := e.Transition(id, "sell", sharia.TransitionInput{})
		if err == nil {
			t.Fatal("expected the guard rule to block the contract sell posting")
		}
		// the underlying ledger error surfaces through the engine; assert the
		// contract did not advance and no receivable was created
		assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", 0)
	})
}
```

Note: `e.Transition` maps a `balance.insufficient`-style ledger error to a
sharia error, but a `*guard.Error` from `Commit` will surface as a generic
commit failure — the test asserts the **effect** (sell blocked, no receivable),
which is the real guarantee. If `commitTransition`'s error mapping swallows the
guard error confusingly, that is acceptable for v1 (the deny still aborted the
posting); note it in the report.

- [ ] **Step 2: Run** `go test ./sharia/ -run TestGuardCoversContractEnginePostings -count=1 -v` → PASS
- [ ] **Step 3: Commit** `git add sharia/guard_coverage_test.go && git commit -m "test(guard): a @contracts:* rule blocks contract-engine postings"`

---

### Task 8: API — rule CRUD + events

**Files:** Create `api/actions/guard_controller.go`; Modify `api/actions/controllers.go`, `api/routes/routes.go`

- [ ] **Step 1: Write `api/actions/guard_controller.go`** following `contract_controller.go` patterns (BaseController, `c.Get("ledger")`, per-ledger). Endpoints:
  - `PostRule` — bind JSON into `guard.Rule`; generate ID if empty (`"rule-" + short random`, reuse a small helper or `fmt.Sprintf("rule-%d", time.Now().UnixNano())`); set timestamps; `guard.ValidateRule(&r)` (400 on error via the guard error's HTTPStatus); `l.Store().SaveRule(r)`; `l.Guard().Reload()`; 201 with the rule.
  - `ListRules` — `l.Store().ListRules()` → 200.
  - `GetRule` — `l.Store().GetRule(id)` → 200 / 404.
  - `PatchRule` — load, apply changed fields (enabled/action/params/reason/standard_ref), re-`ValidateRule`, `UpdateRule`, `Reload` → 200.
  - `DeleteRule` — `DeleteRule(id)`, `Reload` → 200.
  - `ListEvents` — `l.Store().ListGuardEvents(limit, offset)` → 200 (parse `?limit&offset`, default 50/0).
  Render guard errors with their `HTTPStatus()` and the `{error,message,standard_ref?,rule_id?}` shape (mirror the sharia controller's `responseShariaError`).

- [ ] **Step 2: Wire fx + routes.** `api/actions/controllers.go`: add `fx.Provide(NewGuardController)`. `api/routes/routes.go`: add `guardController actions.GuardController` to the struct + `NewRoutes` params + assignment, and inside the `/:ledger` group:

```go
		ledger.POST("/guard/rules", r.guardController.PostRule)
		ledger.GET("/guard/rules", r.guardController.ListRules)
		ledger.GET("/guard/rules/:id", r.guardController.GetRule)
		ledger.PATCH("/guard/rules/:id", r.guardController.PatchRule)
		ledger.DELETE("/guard/rules/:id", r.guardController.DeleteRule)
		ledger.GET("/guard/events", r.guardController.ListEvents)
```

- [ ] **Step 3: httptest** `api/actions/guard_controller_test.go` (reuse `withAPI`/`do`/`doWithHeaders` from `contract_controller_test.go`): create an amount_cap deny rule → 201; create a deny rule WITHOUT standard_ref → 400 `ERR_INVALID_RULE`; list rules → 200; fund a client + post a transaction over the cap via `/transactions` → 422 / blocked; GET `/guard/events` shows the deny; DELETE the rule → 200.

- [ ] **Step 4: Run** `go test ./api/actions/ -count=1` → PASS, then `go test ./... -count=1` → all green.
- [ ] **Step 5: Commit** `git add api/ && git commit -m "feat(api): guard rule CRUD + events endpoints (hot-reload on mutation)"`

---

### Task 9: Docs

**Files:** Create `guard/README.md`

- [ ] One page: what Guard is, the `ledger.Commit` chokepoint (and why not corren-vm), the rule kinds + JSON examples, deny-vs-monitor, the Sharia↔Guard boundary (hard invariants stay in `sharia/`), the durability rule (deny event survives rollback), and the API. Commit `git add guard/README.md && git commit -m "docs(guard): README"`.

---

## Self-review notes

- **Spec coverage:** chokepoint at Commit (T6) ✓; 3 rule kinds (T2) ✓; deny+monitor (T5/T6) ✓; standard_ref required for deny (T3) ✓; cut criterion = txs+netFlows only (matchers take only those, T2) ✓; deny-event survives rollback (T5 unit + T6 keystone) ✓; monitor best-effort post-commit (T6) ✓; RWMutex snapshot + race test (T5) ✓; guard_rules/guard_events v004 both flavors (T4) ✓; scoping covers contract postings (T7) ✓; zero-rule no-op (T6) ✓; API hot-reload (T8) ✓; Sharia↔Guard boundary documented (T9, and structurally: guard rules can't be sharia invariants — different package/store) ✓.
- **Durability mechanics:** verified against `ledger.go` — Guard hook sits after the balance loop, before `SaveTransactions`; deny returns before SaveTransactions and before `l._last` advances, so no state change and no chain advance; the deny event is a separate `AppendGuardEvent` Exec, independent of the (never-run) SaveTransactions.
- **Type consistency:** `Rule`/`GuardEvent`/`Error`/param structs defined in T1, used identically in T2–T8. `Engine.Evaluate(view, txs, netFlows) ([]GuardEvent, error)` and `WriteMonitorEvents([]GuardEvent)` consistent across T5/T6. `GuardStore` methods (T4) match the fake (T5) and the real impls.
- **Import graph:** `guard` imports only `core`; `storage` and `ledger` import `guard`; no cycle.
- **v2 deferred:** velocity, hash-chained events, strict monitor atomicity, Horizon UI — out of scope, noted in spec.

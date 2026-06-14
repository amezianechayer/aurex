package ledger

import (
	"encoding/json"
	"testing"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/guard"
)

func saveRuleAndReload(t *testing.T, l *Ledger, r guard.Rule) {
	t.Helper()
	// The ledger test package shares one sqlite DB across all tests, so a rule
	// saved by an earlier test would otherwise leak into this one. Clear any
	// pre-existing rules first so each guard test sees only the rule it sets.
	existing, err := l.store.ListRules()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range existing {
		if err := l.store.DeleteRule(e.ID); err != nil {
			t.Fatal(err)
		}
	}
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

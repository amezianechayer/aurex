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

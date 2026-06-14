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

func (f *fakeStore) SaveRule(r Rule) error           { f.rules = append(f.rules, r); return nil }
func (f *fakeStore) UpdateRule(r Rule) error         { return nil }
func (f *fakeStore) DeleteRule(id string) error      { return nil }
func (f *fakeStore) GetRule(id string) (Rule, error) { return Rule{}, &Error{Code: ErrNotFound} }
func (f *fakeStore) ListRules() ([]Rule, error)      { return f.rules, nil }
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

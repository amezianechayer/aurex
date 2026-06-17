package flows

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/sharia"
)

// --- fakes ---

type fakeLedger struct {
	scripts  []string
	execErr  error
	accounts map[string]core.Account
}

func (f *fakeLedger) Execute(s core.Script) error {
	f.scripts = append(f.scripts, s.Plain)
	return f.execErr
}
func (f *fakeLedger) GetAccount(addr string) (core.Account, error) {
	if a, ok := f.accounts[addr]; ok {
		return a, nil
	}
	return core.Account{Address: addr, Balances: map[string]int64{}}, nil
}

type fakeSharia struct {
	created     []sharia.CreateRequest
	createErr   error
	transitions []string
}

func (f *fakeSharia) Create(req sharia.CreateRequest) (sharia.Contract, []sharia.Installment, error) {
	f.created = append(f.created, req)
	if f.createErr != nil {
		return sharia.Contract{}, nil, f.createErr
	}
	return sharia.Contract{ID: req.ID, Type: req.Type, State: "PROMISE"}, nil, nil
}
func (f *fakeSharia) Transition(id, name string, in sharia.TransitionInput) (sharia.TransitionResult, error) {
	f.transitions = append(f.transitions, id+":"+name)
	return sharia.TransitionResult{}, nil
}

type fakeStore struct {
	flows     map[string]Flow
	instances map[string]FlowInstance
}

func newFakeStore() *fakeStore {
	return &fakeStore{flows: map[string]Flow{}, instances: map[string]FlowInstance{}}
}
func (s *fakeStore) SaveFlow(f Flow) error   { s.flows[f.ID] = f; return nil }
func (s *fakeStore) UpdateFlow(f Flow) error { s.flows[f.ID] = f; return nil }
func (s *fakeStore) GetFlow(id string) (Flow, error) {
	f, ok := s.flows[id]
	if !ok {
		return Flow{}, fmt.Errorf("not found")
	}
	return f, nil
}
func (s *fakeStore) ListFlows() ([]Flow, error) {
	out := []Flow{}
	for _, f := range s.flows {
		out = append(out, f)
	}
	return out, nil
}
func (s *fakeStore) SaveInstance(i FlowInstance) error   { s.instances[i.ID] = i; return nil }
func (s *fakeStore) UpdateInstance(i FlowInstance) error { s.instances[i.ID] = i; return nil }
func (s *fakeStore) GetInstance(id string) (FlowInstance, error) {
	i, ok := s.instances[id]
	if !ok {
		return FlowInstance{}, fmt.Errorf("not found")
	}
	return i, nil
}
func (s *fakeStore) ListInstances(flowID string) ([]FlowInstance, error) {
	out := []FlowInstance{}
	for _, i := range s.instances {
		if i.FlowID == flowID {
			out = append(out, i)
		}
	}
	return out, nil
}
func (s *fakeStore) ListInstancesByStatus(status string) ([]FlowInstance, error) {
	out := []FlowInstance{}
	for _, i := range s.instances {
		if i.Status == status {
			out = append(out, i)
		}
	}
	return out, nil
}

func newEngine() (*Engine, *fakeLedger, *fakeSharia) {
	led, sh := &fakeLedger{accounts: map[string]core.Account{}}, &fakeSharia{}
	return NewEngine(led, sh, newFakeStore()), led, sh
}

// --- tests ---

func TestManualTransferFlow(t *testing.T) {
	eng, led, _ := newEngine()
	steps := json.RawMessage(`[{"type":"transfer","config":{"script":"transfer [$amount] (from @wallet:$sender to @wallet:$receiver)"}}]`)
	f, err := eng.CreateFlow("p2p_transfer", TriggerManual, "", steps)
	if err != nil {
		t.Fatalf("create flow: %v", err)
	}
	input := json.RawMessage(`{"amount":"AED.2 20000","sender":"alice","receiver":"bob"}`)
	inst, err := eng.Trigger(f.ID, input)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if inst.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s (%s)", inst.Status, inst.Error)
	}
	want := "transfer [AED.2 20000] (from @wallet:alice to @wallet:bob)"
	if len(led.scripts) != 1 || led.scripts[0] != want {
		t.Fatalf("expected substituted script %q, got %v", want, led.scripts)
	}
}

func TestShariaFlow(t *testing.T) {
	eng, _, sh := newEngine()
	steps := json.RawMessage(`[{"type":"sharia","config":{"action":"create","contract_type":"murabaha","params":{"installments":12}}}]`)
	f, _ := eng.CreateFlow("auto_murabaha", TriggerManual, "", steps)
	inst, err := eng.Trigger(f.ID, nil)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if inst.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s (%s)", inst.Status, inst.Error)
	}
	if len(sh.created) != 1 || sh.created[0].Type != "murabaha" {
		t.Fatalf("expected one murabaha created, got %v", sh.created)
	}
}

func TestConditionStopsFlow(t *testing.T) {
	eng, _, sh := newEngine()
	steps := json.RawMessage(`[{"type":"condition","config":{"expr":"amount >= 100000"}},{"type":"sharia","config":{"contract_type":"murabaha","params":{}}}]`)
	f, _ := eng.CreateFlow("auto_murabaha", TriggerManual, "", steps)

	// below threshold → stops, no contract
	inst, _ := eng.Trigger(f.ID, json.RawMessage(`{"amount":50000}`))
	if inst.Status != StatusStopped {
		t.Fatalf("expected stopped, got %s", inst.Status)
	}
	if len(sh.created) != 0 {
		t.Fatalf("no contract should be created when the condition is false")
	}

	// at/above threshold → passes, contract created
	inst, _ = eng.Trigger(f.ID, json.RawMessage(`{"amount":150000}`))
	if inst.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s (%s)", inst.Status, inst.Error)
	}
	if len(sh.created) != 1 {
		t.Fatalf("expected one contract created above threshold, got %d", len(sh.created))
	}
}

func TestWaitThenResume(t *testing.T) {
	eng, led, _ := newEngine()
	steps := json.RawMessage(`[{"type":"transfer","config":{"script":"transfer [DZD.2 1] (from @world to @a)"}},{"type":"wait","config":{"days":7}},{"type":"transfer","config":{"script":"transfer [DZD.2 1] (from @world to @b)"}}]`)
	f, _ := eng.CreateFlow("delayed", TriggerManual, "", steps)

	inst, _ := eng.Trigger(f.ID, nil)
	if inst.Status != StatusWaiting {
		t.Fatalf("expected waiting after the wait step, got %s", inst.Status)
	}
	if len(led.scripts) != 1 {
		t.Fatalf("only the first transfer should have run, got %d", len(led.scripts))
	}
	if inst.ResumeAt == "" {
		t.Fatal("waiting instance must have a resume_at")
	}

	resumed, err := eng.Resume(inst.ID)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if resumed.Status != StatusCompleted {
		t.Fatalf("expected completed after resume, got %s", resumed.Status)
	}
	if len(led.scripts) != 2 {
		t.Fatalf("the second transfer should have run after resume, got %d", len(led.scripts))
	}
}

func TestResumeDue(t *testing.T) {
	eng, _, _ := newEngine()
	steps := json.RawMessage(`[{"type":"wait","config":{"days":1}},{"type":"transfer","config":{"script":"transfer [DZD.2 1] (from @world to @a)"}}]`)
	f, _ := eng.CreateFlow("waiter", TriggerManual, "", steps)
	inst, _ := eng.Trigger(f.ID, nil)
	if inst.Status != StatusWaiting {
		t.Fatalf("expected waiting, got %s", inst.Status)
	}
	// nothing due yet
	if n, _ := eng.ResumeDue(now()); n != 0 {
		t.Fatalf("expected 0 resumed now, got %d", n)
	}
	// 2 days in the future → due
	future := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)
	if n, _ := eng.ResumeDue(future); n != 1 {
		t.Fatalf("expected 1 resumed in the future, got %d", n)
	}
}

func TestStepFailureMarksFailed(t *testing.T) {
	eng, led, _ := newEngine()
	led.execErr = fmt.Errorf("balance.insufficient.DZD.2")
	steps := json.RawMessage(`[{"type":"transfer","config":{"script":"transfer [DZD.2 1] (from @x to @y)"}}]`)
	f, _ := eng.CreateFlow("will_fail", TriggerManual, "", steps)
	inst, err := eng.Trigger(f.ID, nil)
	if err == nil {
		t.Fatal("expected an error from a failing step")
	}
	if inst.Status != StatusFailed || inst.Error == "" {
		t.Fatalf("expected failed instance with an error, got %s / %q", inst.Status, inst.Error)
	}
}

func TestOnWalletCreditedAutoMurabaha(t *testing.T) {
	eng, _, sh := newEngine()
	steps := json.RawMessage(`[{"type":"condition","config":{"expr":"amount >= 100000"}},{"type":"sharia","config":{"contract_type":"murabaha","params":{"client":"$wallet_id"}}}]`)
	if _, err := eng.CreateFlow("auto_murabaha", TriggerWalletCredited, "", steps); err != nil {
		t.Fatalf("create flow: %v", err)
	}

	// a small credit does not trigger the contract
	eng.OnWalletCredited("wlt_alice", "AED.2", 50000)
	if len(sh.created) != 0 {
		t.Fatalf("small credit must not create a contract, got %d", len(sh.created))
	}

	// a large credit does, with the wallet id substituted into the params
	eng.OnWalletCredited("wlt_alice", "AED.2", 150000)
	if len(sh.created) != 1 {
		t.Fatalf("large credit should create one contract, got %d", len(sh.created))
	}
	if got := string(sh.created[0].Params); got != `{"client":"wlt_alice"}` {
		t.Fatalf("expected wallet id substituted into params, got %s", got)
	}
}

func TestCreateFlowRejectsUnknownStep(t *testing.T) {
	eng, _, _ := newEngine()
	_, err := eng.CreateFlow("bad", TriggerManual, "", json.RawMessage(`[{"type":"frobnicate"}]`))
	fe, ok := err.(*Error)
	if !ok || fe.Code != ErrInvalidStep {
		t.Fatalf("expected ERR_INVALID_STEP, got %v", err)
	}
}

func TestCreateFlowValidatesTrigger(t *testing.T) {
	eng, _, _ := newEngine()
	if _, err := eng.CreateFlow("x", "telepathy", "", json.RawMessage(`[{"type":"wait","config":{"days":1}}]`)); err == nil {
		t.Fatal("expected an error for an unknown trigger")
	}
	if _, err := eng.CreateFlow("x", TriggerSchedule, "not-a-schedule", json.RawMessage(`[{"type":"wait","config":{"days":1}}]`)); err == nil {
		t.Fatal("expected an error for an invalid schedule")
	}
}

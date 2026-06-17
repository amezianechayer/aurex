package flows

import (
	"encoding/json"
	"strings"
	"time"
)

// Engine runs flows against one ledger + its sharia engine. It is constructed
// per ledger (corren is multi-ledger), like the other module engines.
type Engine struct {
	led    LedgerPort
	sharia ShariaPort
	store  FlowStore
}

func NewEngine(led LedgerPort, sh ShariaPort, store FlowStore) *Engine {
	return &Engine{led: led, sharia: sh, store: store}
}

// CreateFlow validates and stores a flow definition.
func (e *Engine) CreateFlow(name, trigger, schedule string, steps json.RawMessage) (Flow, error) {
	if name == "" {
		return Flow{}, newError(ErrInvalidParams, "name is required")
	}
	switch trigger {
	case TriggerManual, TriggerWalletCredited, TriggerSchedule:
	default:
		return Flow{}, newError(ErrInvalidParams, "unknown trigger: "+trigger)
	}
	if trigger == TriggerSchedule {
		if _, ok := parseInterval(schedule); !ok {
			return Flow{}, newError(ErrInvalidParams, "schedule must be @hourly|@daily|@weekly|@every <dur>")
		}
	}
	var ss []Step
	if err := json.Unmarshal(steps, &ss); err != nil || len(ss) == 0 {
		return Flow{}, newError(ErrInvalidParams, "steps must be a non-empty array")
	}
	for _, s := range ss {
		switch s.Type {
		case StepTransfer, StepSharia, StepCondition, StepWait:
		default:
			return Flow{}, newError(ErrInvalidStep, "unknown step type: "+s.Type)
		}
	}
	ts := now()
	f := Flow{ID: generateID("flow_"), Name: name, Trigger: trigger, Schedule: schedule, Steps: steps, CreatedAt: ts, UpdatedAt: ts}
	if err := e.store.SaveFlow(f); err != nil {
		return Flow{}, err
	}
	return f, nil
}

func (e *Engine) GetFlow(id string) (Flow, error) {
	f, err := e.store.GetFlow(id)
	if err != nil {
		return Flow{}, newError(ErrNotFound, "flow not found")
	}
	return f, nil
}

func (e *Engine) ListFlows() ([]Flow, error)                         { return e.store.ListFlows() }
func (e *Engine) ListInstances(flowID string) ([]FlowInstance, error) { return e.store.ListInstances(flowID) }

// Trigger starts and runs a new instance of a flow (the manual trigger).
func (e *Engine) Trigger(flowID string, input json.RawMessage) (FlowInstance, error) {
	f, err := e.GetFlow(flowID)
	if err != nil {
		return FlowInstance{}, err
	}
	return e.start(f, input)
}

func (e *Engine) start(f Flow, input json.RawMessage) (FlowInstance, error) {
	ts := now()
	inst := FlowInstance{ID: generateID("fi_"), FlowID: f.ID, Status: StatusRunning, CurrentStep: 0, Input: input, CreatedAt: ts, UpdatedAt: ts}
	if err := e.store.SaveInstance(inst); err != nil {
		return FlowInstance{}, err
	}
	return e.run(f, inst)
}

// run executes steps from inst.CurrentStep until the flow completes, hits a
// wait, is stopped by a condition, or fails — persisting after each step.
func (e *Engine) run(f Flow, inst FlowInstance) (FlowInstance, error) {
	var steps []Step
	if err := json.Unmarshal(f.Steps, &steps); err != nil {
		return e.fail(inst, "invalid steps: "+err.Error())
	}
	input := parseInput(inst.Input)
	for inst.CurrentStep < len(steps) {
		outcome, err := e.execStep(steps[inst.CurrentStep], input)
		if err != nil {
			return e.fail(inst, err.Error())
		}
		switch outcome.kind {
		case outcomeStop:
			inst.Status = StatusStopped
			inst.UpdatedAt = now()
			_ = e.store.UpdateInstance(inst)
			return inst, nil
		case outcomeWait:
			inst.CurrentStep++ // the wait is satisfied once resumed
			inst.Status = StatusWaiting
			inst.ResumeAt = outcome.resumeAt
			inst.UpdatedAt = now()
			_ = e.store.UpdateInstance(inst)
			return inst, nil
		default: // outcomeNext
			inst.CurrentStep++
			inst.UpdatedAt = now()
			_ = e.store.UpdateInstance(inst)
		}
	}
	inst.Status = StatusCompleted
	inst.UpdatedAt = now()
	_ = e.store.UpdateInstance(inst)
	return inst, nil
}

func (e *Engine) fail(inst FlowInstance, msg string) (FlowInstance, error) {
	inst.Status = StatusFailed
	inst.Error = msg
	inst.UpdatedAt = now()
	_ = e.store.UpdateInstance(inst)
	return inst, &Error{Code: ErrStepFailed, Message: msg, FlowID: inst.FlowID}
}

// Resume continues a waiting instance from where it paused.
func (e *Engine) Resume(instanceID string) (FlowInstance, error) {
	inst, err := e.store.GetInstance(instanceID)
	if err != nil {
		return FlowInstance{}, newError(ErrNotFound, "instance not found")
	}
	if inst.Status != StatusWaiting {
		return inst, nil
	}
	f, err := e.store.GetFlow(inst.FlowID)
	if err != nil {
		return FlowInstance{}, newError(ErrNotFound, "flow not found")
	}
	inst.Status = StatusRunning
	inst.ResumeAt = ""
	return e.run(f, inst)
}

// ResumeDue resumes every waiting instance whose ResumeAt has passed.
func (e *Engine) ResumeDue(nowRFC string) (int, error) {
	waiting, err := e.store.ListInstancesByStatus(StatusWaiting)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, inst := range waiting {
		if inst.ResumeAt != "" && inst.ResumeAt <= nowRFC { // RFC3339 UTC sorts lexically
			if _, err := e.Resume(inst.ID); err == nil {
				n++
			}
		}
	}
	return n, nil
}

// OnWalletCredited fires every wallet_credited flow with the credit as input.
// Best-effort: a failing flow records the failure on its instance and never
// disturbs the credit that triggered it.
func (e *Engine) OnWalletCredited(walletID, asset string, amount int64) {
	fs, err := e.store.ListFlows()
	if err != nil {
		return
	}
	for _, f := range fs {
		if f.Trigger != TriggerWalletCredited {
			continue
		}
		input, _ := json.Marshal(map[string]interface{}{"wallet_id": walletID, "asset": asset, "amount": amount})
		_, _ = e.start(f, input)
	}
}

// RunDue triggers scheduled flows that are due and resumes waiting instances.
// Called by the scheduler tick.
func (e *Engine) RunDue(nowRFC string) {
	fs, err := e.store.ListFlows()
	if err == nil {
		for _, f := range fs {
			if f.Trigger == TriggerSchedule && dueNow(f, nowRFC) {
				f.LastRunAt = nowRFC
				_ = e.store.UpdateFlow(f)
				_, _ = e.start(f, nil)
			}
		}
	}
	_, _ = e.ResumeDue(nowRFC)
}

func parseInput(raw json.RawMessage) map[string]interface{} {
	m := map[string]interface{}{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	return m
}

func parseInterval(schedule string) (time.Duration, bool) {
	switch schedule {
	case "@hourly":
		return time.Hour, true
	case "@daily":
		return 24 * time.Hour, true
	case "@weekly":
		return 7 * 24 * time.Hour, true
	}
	if strings.HasPrefix(schedule, "@every ") {
		if d, err := time.ParseDuration(strings.TrimPrefix(schedule, "@every ")); err == nil && d > 0 {
			return d, true
		}
	}
	return 0, false
}

func dueNow(f Flow, nowRFC string) bool {
	d, ok := parseInterval(f.Schedule)
	if !ok {
		return false
	}
	if f.LastRunAt == "" {
		return true
	}
	last, err := time.Parse(time.RFC3339, f.LastRunAt)
	if err != nil {
		return true
	}
	n, err := time.Parse(time.RFC3339, nowRFC)
	if err != nil {
		return false
	}
	return n.Sub(last) >= d
}

package sharia

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/amezianechayer/corren/core"
)

// LedgerPort is the minimal surface the engine needs from *ledger.Ledger.
// Declared here (not imported) to avoid the sharia → ledger → storage →
// sharia import cycle.
type LedgerPort interface {
	Commit(ts []core.Transaction) ([]core.Transaction, error)
	GetAccount(address string) (core.Account, error)
	SaveMeta(targetType string, targetID string, m core.Metadata) error
}

type CreateRequest struct {
	Type   string          `json:"type"`
	ID     string          `json:"id"`
	Params json.RawMessage `json:"params"`
}

type TransitionInput struct {
	Seq         int    `json:"seq"`
	Rebate      *int64 `json:"rebate"`
	Amount      int64  `json:"amount"`
	Destination string `json:"destination"`
}

type TransitionResult struct {
	Contract   Contract `json:"contract"`
	Transition string   `json:"transition"`
	TxID       int64    `json:"tx_id"`
	NewState   string   `json:"new_state"`
}

// keyedMutex serializes concurrent transitions on the same contract.
type keyedMutex struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func (k *keyedMutex) get(key string) *sync.Mutex {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.locks == nil {
		k.locks = map[string]*sync.Mutex{}
	}
	if _, ok := k.locks[key]; !ok {
		k.locks[key] = &sync.Mutex{}
	}
	return k.locks[key]
}

type Engine struct {
	ledger LedgerPort
	store  ShariaStore
	km     keyedMutex
}

func NewEngine(l LedgerPort, s ShariaStore) *Engine {
	return &Engine{ledger: l, store: s}
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// appendAudit chains and persists an audit event, returning its seq.
func (e *Engine) appendAudit(ev AuditEvent) (int, error) {
	if ev.CreatedAt == "" {
		ev.CreatedAt = now()
	}
	return AppendChainedAudit(e.store, ev)
}

// deny journalizes a refusal (a refusal is evidence, not a silent error)
// and decorates the typed error with the audit seq.
func (e *Engine) deny(contractID, transition string, serr *Error) error {
	seq, _ := e.appendAudit(AuditEvent{
		ContractID:  contractID,
		Event:       EventRejected,
		Transition:  transition,
		Decision:    DecisionDenied,
		Reason:      serr.Message,
		StandardRef: serr.StandardRef,
		TxID:        -1,
		Payload:     "{}",
	})
	serr.ContractID = contractID
	serr.AuditSeq = seq
	return serr
}

// Create validates params, builds the schedule and persists the contract
// in its initial state. No ledger posting happens here (spec §5.1). It routes
// generically through the kind registry: the engine holds no contract-specific
// knowledge.
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

	// Re-marshal the decoded+defaulted params so the stored form is canonical.
	raw, _ := json.Marshal(p)
	ts := now()
	c = Contract{
		ID:              req.ID,
		Type:            req.Type,
		State:           kindInitialState(req.Type),
		Params:          raw,
		TemplateVersion: templateVersionFor(req.Type),
		CreatedAt:       ts,
		UpdatedAt:       ts,
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
		ContractID:  c.ID,
		Event:       EventCreated,
		Decision:    DecisionAllowed,
		StandardRef: createStandardRef(req.Type),
		TxID:        -1,
		Payload:     string(raw),
		CreatedAt:   ts,
	}); err != nil {
		return c, nil, err
	}

	return c, schedule, nil
}

// Transition executes one FSM transition following the strict order of
// spec §9: lock, load, sharia gate, FSM check, preconditions, build plan,
// commit, post-commit bookkeeping, audit. The engine holds no contract-specific
// knowledge: it routes through the kind registry by Contract.Type.
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

	// 1. Sharia gate — runs BEFORE the FSM so an ownership/possession
	// violation takes priority over a mere sequencing error (even from PROMISE).
	if gErr := kind.ShariaGate(e.ledger, c, p, ev); gErr != nil {
		if se, ok := gErr.(*Error); ok {
			return res, e.deny(contractID, name, se)
		}
		return res, gErr
	}

	// 2. FSM check
	if !contains(kind.AllowedTransitions(c.State), name) {
		return res, e.deny(contractID, name, &Error{
			Code:    ErrInvalidTransition,
			Message: fmt.Sprintf("transition %q is not allowed from state %s", name, c.State),
		})
	}

	// 3. Load schedule + preconditions
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

	// 4. Build the transition plan
	plan, err := kind.BuildPlan(e.ledger, c, p, sched, ev)
	if err != nil {
		if se, ok := err.(*Error); ok {
			return res, e.deny(contractID, name, se)
		}
		return res, err
	}

	// 5. Commit + generic post-commit bookkeeping
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

// applyPlan commits exactly ONE ledger transaction (invariant I-7) when the
// plan carries postings, then runs the generic, contract-agnostic post-commit
// bookkeeping: transaction metadata, installment marks, contract state, the
// main audit event and any extra audit events. A plan with no postings (e.g.
// cancel from PROMISE) skips the commit but still records state + audit.
func (e *Engine) applyPlan(c Contract, name string, plan TransitionPlan) (TransitionResult, error) {
	var res TransitionResult
	ts := now()
	txID := int64(-1)

	if len(plan.Postings) > 0 {
		tx := core.Transaction{Postings: plan.Postings, Reference: plan.Reference}
		committed, err := e.ledger.Commit([]core.Transaction{tx})
		if err != nil {
			if isDuplicateErr(err) {
				return res, e.deny(c.ID, name, &Error{
					Code:    ErrDuplicate,
					Message: fmt.Sprintf("transition already committed (reference %q)", plan.Reference),
				})
			}
			if strings.HasPrefix(err.Error(), "balance.insufficient") {
				return res, e.deny(c.ID, name, &Error{
					Code:    ErrPrecondition,
					Message: err.Error(),
				})
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
		ContractID:  c.ID,
		Event:       event,
		Transition:  name,
		Decision:    DecisionAllowed,
		StandardRef: plan.StandardRef,
		TxID:        txID,
		Payload:     plan.Payload,
	}); err != nil {
		return res, err
	}
	for _, ex := range plan.ExtraAudit {
		ex.ContractID = c.ID
		// Extra audit events (e.g. the settled event) are tied to the same
		// committed transaction; mirror the original engine's res.TxID.
		ex.TxID = txID
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

// VerifyAudit returns the full audit trail and re-verifies the hash chain.
func (e *Engine) VerifyAudit(contractID string) ([]AuditEvent, bool, error) {
	events, err := e.store.GetAudit(contractID)
	if err != nil {
		return nil, false, err
	}
	return events, VerifyChain(contractID, events), nil
}

// GetContract loads a contract with its schedule.
func (e *Engine) GetContract(contractID string) (Contract, []Installment, error) {
	c, err := e.store.GetContract(contractID)
	if err != nil {
		return c, nil, &Error{Code: ErrNotFound, Message: "contract not found", ContractID: contractID}
	}
	schedule, err := e.store.GetSchedule(contractID)
	return c, schedule, err
}

func isDuplicateErr(err error) bool {
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unique") || strings.Contains(s, "duplicate")
}

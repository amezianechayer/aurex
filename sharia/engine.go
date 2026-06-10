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
	Type   string         `json:"type"`
	ID     string         `json:"id"`
	Params MurabahaParams `json:"params"`
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
// in state PROMISE. No ledger posting happens here (spec §5.1).
func (e *Engine) Create(req CreateRequest) (Contract, []Installment, error) {
	var c Contract

	if req.Type == "" {
		req.Type = TypeMurabaha
	}
	if req.Type != TypeMurabaha {
		return c, nil, newError(ErrInvalidParams, "type must be \"murabaha\" in v1")
	}

	if req.ID == "" {
		req.ID = GenerateContractID()
	}
	if !ContractIDIsValid(req.ID) {
		return c, nil, newError(ErrInvalidParams, "id must match ^[a-z][a-z0-9_]{3,40}$")
	}

	if err := req.Params.Validate(); err != nil {
		return c, nil, err
	}

	if _, err := e.store.GetContract(req.ID); err == nil {
		return c, nil, &Error{Code: ErrDuplicate, Message: "contract already exists", ContractID: req.ID}
	}

	schedule, err := BuildSchedule(
		req.Params.Cost.Amount,
		req.Params.Markup.Amount,
		req.Params.Installments,
		req.Params.FirstDue,
		req.Params.PeriodDays,
	)
	if err != nil {
		return c, nil, newError(ErrInvalidParams, err.Error())
	}

	ts := now()
	c = Contract{
		ID:              req.ID,
		Type:            TypeMurabaha,
		State:           StatePromise,
		Params:          req.Params,
		TemplateVersion: TemplateVersionMurabaha,
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

	payload, _ := json.Marshal(req.Params)
	if _, err := e.appendAudit(AuditEvent{
		ContractID:  c.ID,
		Event:       EventCreated,
		Decision:    DecisionAllowed,
		StandardRef: RefSS8,
		TxID:        -1,
		Payload:     string(payload),
		CreatedAt:   ts,
	}); err != nil {
		return c, nil, err
	}

	return c, schedule, nil
}

// Transition executes one FSM transition following the strict order of
// spec §9: lock, load, FSM check, preconditions, postings, commit,
// post-commit bookkeeping, audit.
func (e *Engine) Transition(contractID, name string, input TransitionInput) (TransitionResult, error) {
	var res TransitionResult

	lock := e.km.get(contractID)
	lock.Lock()
	defer lock.Unlock()

	c, err := e.store.GetContract(contractID)
	if err != nil {
		return res, &Error{Code: ErrNotFound, Message: "contract not found", ContractID: contractID}
	}

	// Spec §5.3 orders the sell preconditions: 1. qabd (I-1), 2. state.
	// Selling a non-possessed asset is a Sharia violation, not a mere
	// sequencing error — even from PROMISE.
	if name == TransitionSell {
		inventory, err := e.ledger.GetAccount(InventoryAccount(c.ID))
		if err != nil {
			return res, err
		}
		if inventory.Balances[c.Params.AssetCode] < 1 {
			return res, e.deny(contractID, name, &Error{
				Code:        ErrShariaViolation,
				Message:     "sale of non-possessed asset",
				StandardRef: RefSS8,
			})
		}
	}

	if !TransitionAllowed(c.State, name) {
		return res, e.deny(contractID, name, &Error{
			Code:    ErrInvalidTransition,
			Message: fmt.Sprintf("transition %q is not allowed from state %s", name, c.State),
		})
	}

	switch name {
	case TransitionAcquire:
		return e.acquire(c)
	case TransitionSell:
		return e.sell(c)
	case TransitionPayInstallment:
		return e.payInstallment(c, input)
	case TransitionEarlySettle:
		return e.earlySettle(c, input)
	case TransitionLatePenalty:
		return e.latePenalty(c, input)
	case TransitionCancel:
		return e.cancel(c)
	}

	return res, e.deny(contractID, name, &Error{
		Code:    ErrInvalidTransition,
		Message: fmt.Sprintf("unknown transition %q", name),
	})
}

func (e *Engine) acquire(c Contract) (TransitionResult, error) {
	var res TransitionResult
	p := c.Params

	treasury, err := e.ledger.GetAccount(p.BankTreasury)
	if err != nil {
		return res, err
	}
	if treasury.Balances[p.Cost.Asset] < p.Cost.Amount {
		return res, e.deny(c.ID, TransitionAcquire, &Error{
			Code:    ErrPrecondition,
			Message: "insufficient treasury balance to acquire the asset",
		})
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"cost": p.Cost, "supplier": p.Supplier,
	})
	return e.commitTransition(c, TransitionAcquire, commitSpec{
		postings:  AcquirePostings(c.ID, p),
		reference: c.ID + ":acquire",
		newState:  StateAcquired,
		ref:       RefSS8,
		payload:   string(payload),
	})
}

func (e *Engine) sell(c Contract) (TransitionResult, error) {
	var res TransitionResult
	p := c.Params

	// Invariant I-1 (qabd): the bank must possess the asset before selling.
	inventory, err := e.ledger.GetAccount(InventoryAccount(c.ID))
	if err != nil {
		return res, err
	}
	if inventory.Balances[p.AssetCode] < 1 {
		return res, e.deny(c.ID, TransitionSell, &Error{
			Code:        ErrShariaViolation,
			Message:     "sale of non-possessed asset",
			StandardRef: RefSS8,
		})
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"total":  p.Cost.Amount + p.Markup.Amount,
		"markup": p.Markup,
		"client": p.Client,
	})
	return e.commitTransition(c, TransitionSell, commitSpec{
		postings:  SellPostings(c.ID, p),
		reference: c.ID + ":sell",
		newState:  StateSold,
		ref:       RefSS8,
		payload:   string(payload),
	})
}

func (e *Engine) payInstallment(c Contract, input TransitionInput) (TransitionResult, error) {
	var res TransitionResult
	p := c.Params

	schedule, err := e.store.GetSchedule(c.ID)
	if err != nil {
		return res, err
	}

	var target *Installment
	if input.Seq > 0 {
		for i := range schedule {
			if schedule[i].Seq == input.Seq {
				target = &schedule[i]
				break
			}
		}
		if target == nil {
			return res, e.deny(c.ID, TransitionPayInstallment, &Error{
				Code:    ErrNotFound,
				Message: fmt.Sprintf("installment %d does not exist", input.Seq),
			})
		}
	} else {
		for i := range schedule {
			if schedule[i].Status == StatusPending || schedule[i].Status == StatusOverdue {
				target = &schedule[i]
				break
			}
		}
		if target == nil {
			return res, e.deny(c.ID, TransitionPayInstallment, &Error{
				Code:    ErrPrecondition,
				Message: "no unpaid installment left",
			})
		}
	}

	// Replaying the payment of an already-paid installment is the
	// idempotence case (scenario E): same reference, clean failure.
	if target.Status == StatusPaid {
		return res, e.deny(c.ID, TransitionPayInstallment, &Error{
			Code:    ErrDuplicate,
			Message: fmt.Sprintf("installment %d is already paid", target.Seq),
		})
	}
	if target.Status == StatusSettledEarly {
		return res, e.deny(c.ID, TransitionPayInstallment, &Error{
			Code:    ErrPrecondition,
			Message: fmt.Sprintf("installment %d was settled early", target.Seq),
		})
	}

	client, err := e.ledger.GetAccount(p.Client)
	if err != nil {
		return res, err
	}
	if client.Balances[p.Cost.Asset] < target.Amount {
		return res, e.deny(c.ID, TransitionPayInstallment, &Error{
			Code:    ErrPrecondition,
			Message: "insufficient client balance",
		})
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"seq": target.Seq, "amount": target.Amount, "profit_part": target.ProfitPart,
	})
	res, err = e.commitTransition(c, TransitionPayInstallment, commitSpec{
		postings:  PayInstallmentPostings(c.ID, p, *target),
		reference: fmt.Sprintf("%s:pay:%d", c.ID, target.Seq),
		newState:  StateSold,
		ref:       RefFAS28,
		payload:   string(payload),
		// state update deferred: it depends on whether this was the last one
		skipStateUpdate: true,
	})
	if err != nil {
		return res, err
	}

	if err := e.store.MarkInstallment(c.ID, target.Seq, StatusPaid, res.TxID, now()); err != nil {
		return res, err
	}

	// settle if every installment is now paid or settled early
	settled := true
	for _, it := range schedule {
		if it.Seq == target.Seq {
			continue
		}
		if it.Status == StatusPending || it.Status == StatusOverdue {
			settled = false
			break
		}
	}

	newState := StateSold
	if settled {
		newState = StateSettled
		if _, err := e.appendAudit(AuditEvent{
			ContractID:  c.ID,
			Event:       EventSettled,
			Decision:    DecisionAllowed,
			StandardRef: RefFAS28,
			TxID:        res.TxID,
			Payload:     "{}",
		}); err != nil {
			return res, err
		}
	}
	if err := e.store.UpdateContractState(c.ID, newState, now()); err != nil {
		return res, err
	}

	res.NewState = newState
	res.Contract.State = newState
	return res, nil
}

func (e *Engine) earlySettle(c Contract, input TransitionInput) (TransitionResult, error) {
	var res TransitionResult
	p := c.Params

	schedule, err := e.store.GetSchedule(c.ID)
	if err != nil {
		return res, err
	}

	var restTotal, restProfit int64
	rest := []Installment{}
	for _, it := range schedule {
		if it.Status == StatusPending || it.Status == StatusOverdue {
			restTotal += it.Amount
			restProfit += it.ProfitPart
			rest = append(rest, it)
		}
	}
	if len(rest) == 0 {
		return res, e.deny(c.ID, TransitionEarlySettle, &Error{
			Code:    ErrPrecondition,
			Message: "no unpaid installment left",
		})
	}

	// ibra': default is a full rebate of the unearned profit — at the
	// bank's discretion, never stipulated in advance as an obligation.
	rebate := restProfit
	if input.Rebate != nil {
		rebate = *input.Rebate
	}
	if rebate < 0 || rebate > restProfit {
		return res, e.deny(c.ID, TransitionEarlySettle, &Error{
			Code:    ErrInvalidParams,
			Message: fmt.Sprintf("rebate must be between 0 and the remaining profit (%d)", restProfit),
		})
	}

	cashDue := restTotal - rebate
	client, err := e.ledger.GetAccount(p.Client)
	if err != nil {
		return res, err
	}
	if client.Balances[p.Cost.Asset] < cashDue {
		return res, e.deny(c.ID, TransitionEarlySettle, &Error{
			Code:    ErrPrecondition,
			Message: "insufficient client balance",
		})
	}

	payload, _ := json.Marshal(map[string]int64{
		"rest_total": restTotal, "rest_profit": restProfit, "rebate": rebate, "cash_due": cashDue,
	})
	res, err = e.commitTransition(c, TransitionEarlySettle, commitSpec{
		postings:        EarlySettlePostings(c.ID, p, restTotal, restProfit, rebate),
		reference:       c.ID + ":early_settle",
		newState:        StateSettled,
		ref:             RefFAS28,
		payload:         string(payload),
		skipStateUpdate: true,
	})
	if err != nil {
		return res, err
	}

	ts := now()
	for _, it := range rest {
		if err := e.store.MarkInstallment(c.ID, it.Seq, StatusSettledEarly, res.TxID, ts); err != nil {
			return res, err
		}
	}
	if _, err := e.appendAudit(AuditEvent{
		ContractID:  c.ID,
		Event:       EventSettled,
		Decision:    DecisionAllowed,
		StandardRef: RefFAS28,
		TxID:        res.TxID,
		Payload:     string(payload),
	}); err != nil {
		return res, err
	}
	if err := e.store.UpdateContractState(c.ID, StateSettled, ts); err != nil {
		return res, err
	}

	res.NewState = StateSettled
	res.Contract.State = StateSettled
	return res, nil
}

func (e *Engine) latePenalty(c Contract, input TransitionInput) (TransitionResult, error) {
	var res TransitionResult
	p := c.Params

	if input.Amount <= 0 {
		return res, e.deny(c.ID, TransitionLatePenalty, &Error{
			Code:    ErrInvalidParams,
			Message: "penalty amount must be > 0",
		})
	}

	destination := input.Destination
	if destination == "" {
		destination = DefaultCharityPool
	}

	// Invariant I-3: penalties may only ever flow to charity (AAOIFI SS 3).
	postings, err := PenaltyPostings(c.ID, p, input.Amount, destination)
	if err != nil {
		return res, e.deny(c.ID, TransitionLatePenalty, err.(*Error))
	}

	if input.Seq > 0 {
		schedule, err := e.store.GetSchedule(c.ID)
		if err != nil {
			return res, err
		}
		found := false
		for _, it := range schedule {
			if it.Seq == input.Seq {
				found = true
				break
			}
		}
		if !found {
			return res, e.deny(c.ID, TransitionLatePenalty, &Error{
				Code:    ErrNotFound,
				Message: fmt.Sprintf("installment %d does not exist", input.Seq),
			})
		}
	}

	client, err := e.ledger.GetAccount(p.Client)
	if err != nil {
		return res, err
	}
	if client.Balances[p.Cost.Asset] < input.Amount {
		return res, e.deny(c.ID, TransitionLatePenalty, &Error{
			Code:    ErrPrecondition,
			Message: "insufficient client balance",
		})
	}

	reference := c.ID + ":late_penalty"
	if input.Seq > 0 {
		reference = fmt.Sprintf("%s:late_penalty:%d", c.ID, input.Seq)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"seq": input.Seq, "amount": input.Amount, "destination": destination,
	})
	return e.commitTransition(c, TransitionLatePenalty, commitSpec{
		postings:  postings,
		reference: reference,
		newState:  c.State, // a penalty never changes the contract state
		ref:       RefSS3,
		payload:   string(payload),
		event:     EventPenalty,
	})
}

func (e *Engine) cancel(c Contract) (TransitionResult, error) {
	postings := CancelPostings(c.ID, c.Params, c.State)

	if len(postings) == 0 {
		// from PROMISE: nothing has moved, no ledger transaction
		ts := now()
		if err := e.store.UpdateContractState(c.ID, StateCancelled, ts); err != nil {
			return TransitionResult{}, err
		}
		if _, err := e.appendAudit(AuditEvent{
			ContractID: c.ID,
			Event:      EventTransition,
			Transition: TransitionCancel,
			Decision:   DecisionAllowed,
			TxID:       -1,
			Payload:    "{}",
		}); err != nil {
			return TransitionResult{}, err
		}
		c.State = StateCancelled
		c.UpdatedAt = ts
		return TransitionResult{Contract: c, Transition: TransitionCancel, TxID: -1, NewState: StateCancelled}, nil
	}

	return e.commitTransition(c, TransitionCancel, commitSpec{
		postings:  postings,
		reference: c.ID + ":cancel",
		newState:  StateCancelled,
		payload:   "{}",
	})
}

type commitSpec struct {
	postings        []core.Posting
	reference       string
	newState        string
	ref             string // AAOIFI standard reference
	payload         string
	event           string // defaults to EventTransition
	skipStateUpdate bool
}

// commitTransition commits exactly ONE ledger transaction (invariant I-7),
// then runs the post-commit bookkeeping: transaction metadata, contract
// state, audit. If the process crashes between commit and bookkeeping, the
// unique Reference allows reconciliation on restart — kept deliberately
// simple in v1, no extra mechanism.
func (e *Engine) commitTransition(c Contract, name string, spec commitSpec) (TransitionResult, error) {
	var res TransitionResult

	tx := core.Transaction{
		Postings:  spec.postings,
		Reference: spec.reference,
	}
	committed, err := e.ledger.Commit([]core.Transaction{tx})
	if err != nil {
		if isDuplicateErr(err) {
			return res, e.deny(c.ID, name, &Error{
				Code:    ErrDuplicate,
				Message: fmt.Sprintf("transition already committed (reference %q)", spec.reference),
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
	txID := committed[0].ID

	e.ledger.SaveMeta("transaction", fmt.Sprint(txID), core.Metadata{
		"sharia/contract": json.RawMessage(fmt.Sprintf("%q", c.ID)),
		"sharia/event":    json.RawMessage(fmt.Sprintf("%q", name)),
		"sharia/ref":      json.RawMessage(fmt.Sprintf("%q", spec.ref)),
	})

	ts := now()
	if !spec.skipStateUpdate && spec.newState != c.State {
		if err := e.store.UpdateContractState(c.ID, spec.newState, ts); err != nil {
			return res, err
		}
	}

	event := spec.event
	if event == "" {
		event = EventTransition
	}
	if _, err := e.appendAudit(AuditEvent{
		ContractID:  c.ID,
		Event:       event,
		Transition:  name,
		Decision:    DecisionAllowed,
		StandardRef: spec.ref,
		TxID:        txID,
		Payload:     spec.payload,
	}); err != nil {
		return res, err
	}

	c.State = spec.newState
	c.UpdatedAt = ts
	return TransitionResult{
		Contract:   c,
		Transition: name,
		TxID:       txID,
		NewState:   spec.newState,
	}, nil
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

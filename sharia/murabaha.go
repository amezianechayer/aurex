package sharia

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/amezianechayer/corren/core"
)

// Transition names
const (
	TransitionAcquire        = "acquire"
	TransitionSell           = "sell"
	TransitionPayInstallment = "pay_installment"
	TransitionEarlySettle    = "early_settle"
	TransitionLatePenalty    = "late_penalty"
	TransitionCancel         = "cancel"
)

// murabahaFSM defines the only executable transitions (invariant I-5).
// cancel from SOLD is deliberately absent (rule C-2).
var murabahaFSM = map[string]map[string]bool{
	StatePromise: {
		TransitionAcquire: true,
		TransitionCancel:  true,
	},
	StateAcquired: {
		TransitionSell:   true,
		TransitionCancel: true,
	},
	StateSold: {
		TransitionPayInstallment: true,
		TransitionEarlySettle:    true,
		TransitionLatePenalty:    true,
	},
}

func TransitionAllowed(state, name string) bool {
	return murabahaFSM[state][name]
}

// AcquirePostings — PROMISE → ACQUIRED (spec §5.2).
// P1 pays the supplier; P2 brings the physical asset into the ledger
// perimeter (the outside world is the source of real goods).
func AcquirePostings(id string, p MurabahaParams) []core.Posting {
	return []core.Posting{
		{Source: p.BankTreasury, Destination: p.Supplier, Asset: p.Cost.Asset, Amount: p.Cost.Amount},
		{Source: core.WORLD, Destination: InventoryAccount(id), Asset: p.AssetCode, Amount: 1},
	}
}

// SellPostings — ACQUIRED → SOLD (spec §5.3).
// The asset moves to the client, the full receivable is born and the
// markup is parked as deferred profit (FAS 28, invariant I-4).
func SellPostings(id string, p MurabahaParams) []core.Posting {
	total := p.Cost.Amount + p.Markup.Amount
	postings := []core.Posting{
		{Source: InventoryAccount(id), Destination: p.Client + ":assets", Asset: p.AssetCode, Amount: 1},
		{Source: CounterpartAccount(id), Destination: ReceivableAccount(id), Asset: p.Cost.Asset, Amount: total},
	}
	if p.Markup.Amount > 0 {
		postings = append(postings, core.Posting{
			Source: CounterpartAccount(id), Destination: DeferredAccount(id), Asset: p.Cost.Asset, Amount: p.Markup.Amount,
		})
	}
	return postings
}

// PayInstallmentPostings — SOLD → SOLD/SETTLED (spec §5.4).
// P3 recognizes only THIS installment's profit share (invariant I-4).
func PayInstallmentPostings(id string, p MurabahaParams, inst Installment) []core.Posting {
	postings := []core.Posting{
		{Source: p.Client, Destination: p.BankTreasury, Asset: p.Cost.Asset, Amount: inst.Amount},
		{Source: ReceivableAccount(id), Destination: CounterpartAccount(id), Asset: p.Cost.Asset, Amount: inst.Amount},
	}
	if inst.ProfitPart > 0 {
		postings = append(postings, core.Posting{
			Source: DeferredAccount(id), Destination: IncomeAccount, Asset: p.Cost.Asset, Amount: inst.ProfitPart,
		})
	}
	return postings
}

// EarlySettlePostings — SOLD → SETTLED with ibra' (spec §5.5).
// The rebated profit is never recognized as income: it is cancelled (P4).
func EarlySettlePostings(id string, p MurabahaParams, restTotal, restProfit, rebate int64) []core.Posting {
	cashDue := restTotal - rebate
	postings := []core.Posting{
		{Source: p.Client, Destination: p.BankTreasury, Asset: p.Cost.Asset, Amount: cashDue},
		{Source: ReceivableAccount(id), Destination: CounterpartAccount(id), Asset: p.Cost.Asset, Amount: restTotal},
	}
	if restProfit-rebate > 0 {
		postings = append(postings, core.Posting{
			Source: DeferredAccount(id), Destination: IncomeAccount, Asset: p.Cost.Asset, Amount: restProfit - rebate,
		})
	}
	if rebate > 0 {
		postings = append(postings, core.Posting{
			Source: DeferredAccount(id), Destination: CounterpartAccount(id), Asset: p.Cost.Asset, Amount: rebate,
		})
	}
	return postings
}

// PenaltyPostings — late penalty (spec §5.6). Hard rule I-3: the only
// lawful destination is a charity account; penalties can never become
// bank income (AAOIFI SS 3).
func PenaltyPostings(id string, p MurabahaParams, amount int64, destination string) ([]core.Posting, error) {
	if !strings.HasPrefix(destination, CharityPrefix) {
		return nil, &Error{
			Code:        ErrShariaViolation,
			Message:     "late penalty destination must be a @charity: account",
			StandardRef: RefSS3,
			ContractID:  id,
		}
	}
	return []core.Posting{
		{Source: p.Client, Destination: destination, Asset: p.Cost.Asset, Amount: amount},
	}, nil
}

// CancelPostings — PROMISE/ACQUIRED → CANCELLED (spec §5.7).
// From ACQUIRED the bank keeps the asset on its books: that ownership
// risk is precisely what legitimates the markup.
func CancelPostings(id string, p MurabahaParams, fromState string) []core.Posting {
	if fromState != StateAcquired {
		return nil
	}
	return []core.Posting{
		{Source: InventoryAccount(id), Destination: UnsoldInventory, Asset: p.AssetCode, Amount: 1},
	}
}

// murabahaKind is the ContractKind implementation for Murabaha. It carries all
// the Murabaha-specific knowledge the engine used to hold inline.
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

// ShariaGate preserves today's behavior: selling a non-possessed asset is an
// SS-8 violation, checked before the FSM even from PROMISE (spec §5.3, I-1).
func (murabahaKind) ShariaGate(led LedgerPort, c Contract, p Params, ev Event) error {
	if ev.Name != TransitionSell {
		return nil
	}
	mp := p.(MurabahaParams)
	inventory, err := led.GetAccount(InventoryAccount(c.ID))
	if err != nil {
		return err
	}
	if inventory.Balances[mp.AssetCode] < 1 {
		return &Error{
			Code:        ErrShariaViolation,
			Message:     "sale of non-possessed asset",
			StandardRef: RefSS8,
		}
	}
	return nil
}

// Preconditions runs the balance/sequence checks that precede the commit for
// each transition, returning the same typed *Error as the old engine methods.
func (murabahaKind) Preconditions(led LedgerPort, c Contract, p Params, sched []Installment, ev Event) error {
	mp := p.(MurabahaParams)
	switch ev.Name {
	case TransitionAcquire:
		treasury, err := led.GetAccount(mp.BankTreasury)
		if err != nil {
			return err
		}
		if treasury.Balances[mp.Cost.Asset] < mp.Cost.Amount {
			return &Error{
				Code:    ErrPrecondition,
				Message: "insufficient treasury balance to acquire the asset",
			}
		}
	case TransitionSell:
		// qabd is enforced in ShariaGate; nothing else to check here.
	case TransitionPayInstallment:
		target := murabahaTarget(sched, ev.Input.Seq)
		if target == nil {
			if ev.Input.Seq > 0 {
				return &Error{
					Code:    ErrNotFound,
					Message: fmt.Sprintf("installment %d does not exist", ev.Input.Seq),
				}
			}
			return &Error{
				Code:    ErrPrecondition,
				Message: "no unpaid installment left",
			}
		}
		// Replaying the payment of an already-paid installment is the
		// idempotence case (scenario E): same reference, clean failure.
		if target.Status == StatusPaid {
			return &Error{
				Code:    ErrDuplicate,
				Message: fmt.Sprintf("installment %d is already paid", target.Seq),
			}
		}
		if target.Status == StatusSettledEarly {
			return &Error{
				Code:    ErrPrecondition,
				Message: fmt.Sprintf("installment %d was settled early", target.Seq),
			}
		}
		client, err := led.GetAccount(mp.Client)
		if err != nil {
			return err
		}
		if client.Balances[mp.Cost.Asset] < target.Amount {
			return &Error{
				Code:    ErrPrecondition,
				Message: "insufficient client balance",
			}
		}
	case TransitionEarlySettle:
		var restTotal, restProfit int64
		rest := []Installment{}
		for _, it := range sched {
			if it.Status == StatusPending || it.Status == StatusOverdue {
				restTotal += it.Amount
				restProfit += it.ProfitPart
				rest = append(rest, it)
			}
		}
		if len(rest) == 0 {
			return &Error{
				Code:    ErrPrecondition,
				Message: "no unpaid installment left",
			}
		}
		// ibra': default is a full rebate of the unearned profit.
		rebate := restProfit
		if ev.Input.Rebate != nil {
			rebate = *ev.Input.Rebate
		}
		if rebate < 0 || rebate > restProfit {
			return &Error{
				Code:    ErrInvalidParams,
				Message: fmt.Sprintf("rebate must be between 0 and the remaining profit (%d)", restProfit),
			}
		}
		cashDue := restTotal - rebate
		client, err := led.GetAccount(mp.Client)
		if err != nil {
			return err
		}
		if client.Balances[mp.Cost.Asset] < cashDue {
			return &Error{
				Code:    ErrPrecondition,
				Message: "insufficient client balance",
			}
		}
	case TransitionLatePenalty:
		if ev.Input.Amount <= 0 {
			return &Error{
				Code:    ErrInvalidParams,
				Message: "penalty amount must be > 0",
			}
		}
		// Charity gate (I-3) runs before seq/balance checks to match the
		// original engine ordering: a non-charity destination is rejected
		// even when the seq is missing or the client is underfunded.
		destination := ev.Input.Destination
		if destination == "" {
			destination = DefaultCharityPool
		}
		if _, err := PenaltyPostings(c.ID, mp, ev.Input.Amount, destination); err != nil {
			return err
		}
		if ev.Input.Seq > 0 {
			found := false
			for _, it := range sched {
				if it.Seq == ev.Input.Seq {
					found = true
					break
				}
			}
			if !found {
				return &Error{
					Code:    ErrNotFound,
					Message: fmt.Sprintf("installment %d does not exist", ev.Input.Seq),
				}
			}
		}
		client, err := led.GetAccount(mp.Client)
		if err != nil {
			return err
		}
		if client.Balances[mp.Cost.Asset] < ev.Input.Amount {
			return &Error{
				Code:    ErrPrecondition,
				Message: "insufficient client balance",
			}
		}
	}
	return nil
}

// BuildPlan produces the postings + state + audit bookkeeping for a transition,
// matching the old engine methods byte-for-byte.
func (murabahaKind) BuildPlan(led LedgerPort, c Contract, p Params, sched []Installment, ev Event) (TransitionPlan, error) {
	mp := p.(MurabahaParams)
	switch ev.Name {
	case TransitionAcquire:
		payload, _ := json.Marshal(map[string]interface{}{
			"cost": mp.Cost, "supplier": mp.Supplier,
		})
		return TransitionPlan{
			Postings:    AcquirePostings(c.ID, mp),
			Reference:   c.ID + ":acquire",
			NewState:    StateAcquired,
			StandardRef: RefSS8,
			Payload:     string(payload),
		}, nil

	case TransitionSell:
		payload, _ := json.Marshal(map[string]interface{}{
			"total":  mp.Cost.Amount + mp.Markup.Amount,
			"markup": mp.Markup,
			"client": mp.Client,
		})
		return TransitionPlan{
			Postings:    SellPostings(c.ID, mp),
			Reference:   c.ID + ":sell",
			NewState:    StateSold,
			StandardRef: RefSS8,
			Payload:     string(payload),
		}, nil

	case TransitionPayInstallment:
		target := murabahaTarget(sched, ev.Input.Seq)
		payload, _ := json.Marshal(map[string]interface{}{
			"seq": target.Seq, "amount": target.Amount, "profit_part": target.ProfitPart,
		})
		newState := StateSold
		var extra []AuditEvent
		if murabahaIsLastUnpaid(sched, target.Seq) {
			newState = StateSettled
			extra = []AuditEvent{{
				Event:       EventSettled,
				Decision:    DecisionAllowed,
				StandardRef: RefFAS28,
				TxID:        -1,
				Payload:     "{}",
			}}
		}
		return TransitionPlan{
			Postings:    PayInstallmentPostings(c.ID, mp, *target),
			Reference:   fmt.Sprintf("%s:pay:%d", c.ID, target.Seq),
			NewState:    newState,
			StandardRef: RefFAS28,
			Payload:     string(payload),
			Marks:       []InstallmentMark{{Seq: target.Seq, Status: StatusPaid}},
			ExtraAudit:  extra,
		}, nil

	case TransitionEarlySettle:
		var restTotal, restProfit int64
		rest := []Installment{}
		for _, it := range sched {
			if it.Status == StatusPending || it.Status == StatusOverdue {
				restTotal += it.Amount
				restProfit += it.ProfitPart
				rest = append(rest, it)
			}
		}
		rebate := restProfit
		if ev.Input.Rebate != nil {
			rebate = *ev.Input.Rebate
		}
		cashDue := restTotal - rebate
		payload, _ := json.Marshal(map[string]int64{
			"rest_total": restTotal, "rest_profit": restProfit, "rebate": rebate, "cash_due": cashDue,
		})
		marks := make([]InstallmentMark, 0, len(rest))
		for _, it := range rest {
			marks = append(marks, InstallmentMark{Seq: it.Seq, Status: StatusSettledEarly})
		}
		return TransitionPlan{
			Postings:    EarlySettlePostings(c.ID, mp, restTotal, restProfit, rebate),
			Reference:   c.ID + ":early_settle",
			NewState:    StateSettled,
			StandardRef: RefFAS28,
			Payload:     string(payload),
			Marks:       marks,
			ExtraAudit: []AuditEvent{{
				Event:       EventSettled,
				Decision:    DecisionAllowed,
				StandardRef: RefFAS28,
				TxID:        -1,
				Payload:     string(payload),
			}},
		}, nil

	case TransitionLatePenalty:
		destination := ev.Input.Destination
		if destination == "" {
			destination = DefaultCharityPool
		}
		postings, err := PenaltyPostings(c.ID, mp, ev.Input.Amount, destination)
		if err != nil {
			return TransitionPlan{}, err
		}
		reference := c.ID + ":late_penalty"
		if ev.Input.Seq > 0 {
			reference = fmt.Sprintf("%s:late_penalty:%d", c.ID, ev.Input.Seq)
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"seq": ev.Input.Seq, "amount": ev.Input.Amount, "destination": destination,
		})
		return TransitionPlan{
			Postings:    postings,
			Reference:   reference,
			NewState:    c.State, // a penalty never changes the contract state
			StandardRef: RefSS3,
			Event:       EventPenalty,
			Payload:     string(payload),
		}, nil

	case TransitionCancel:
		return TransitionPlan{
			Postings:  CancelPostings(c.ID, mp, c.State),
			Reference: c.ID + ":cancel",
			NewState:  StateCancelled,
			Payload:   "{}",
		}, nil
	}

	return TransitionPlan{}, &Error{
		Code:    ErrInvalidTransition,
		Message: fmt.Sprintf("unknown transition %q", ev.Name),
	}
}

// murabahaTarget resolves the installment a pay_installment targets: the
// explicit seq when given, otherwise the first pending/overdue one. Returns nil
// when no match exists.
func murabahaTarget(sched []Installment, seq int) *Installment {
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

// murabahaIsLastUnpaid reports whether paying seq settles the whole schedule:
// every other installment is already paid or settled early.
func murabahaIsLastUnpaid(sched []Installment, seq int) bool {
	for _, it := range sched {
		if it.Seq == seq {
			continue
		}
		if it.Status == StatusPending || it.Status == StatusOverdue {
			return false
		}
	}
	return true
}

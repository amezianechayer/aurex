package sharia

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/amezianechayer/corren/core"
)

const (
	TransitionLease   = "lease"
	TransitionPayRent = "pay_rent"
)

// ijarahFSM defines the only executable transitions for operating Ijarah.
// cancel from LEASED is deliberately absent (a started lease is not undone
// by cancel in v1; early termination is v2).
var ijarahFSM = map[string]map[string]bool{
	StatePromise:  {TransitionAcquire: true, TransitionCancel: true},
	StateAcquired: {TransitionLease: true, TransitionCancel: true},
	StateLeased:   {TransitionPayRent: true, TransitionLatePenalty: true},
}

// IjarahAcquirePostings — PROMISE → ACQUIRED. The bank buys the asset:
// pays the supplier, takes the physical unit, and capitalizes the asset's
// book value on its own books (the lessor keeps ownership throughout).
func IjarahAcquirePostings(id string, p IjarahParams) []core.Posting {
	return []core.Posting{
		{Source: p.BankTreasury, Destination: p.Supplier, Asset: p.Cost.Asset, Amount: p.Cost.Amount},
		{Source: core.WORLD, Destination: AssetAccount(id), Asset: p.AssetCode, Amount: 1},
		{Source: core.WORLD, Destination: AssetAccount(id), Asset: p.Cost.Asset, Amount: p.Cost.Amount},
	}
}

// IjarahLeasePostings — ACQUIRED → LEASED. Only POSSESSION (usufruct) of the
// unit passes to the lessee; the book value stays on the lessor's books and
// no rent receivable is created (rent is recognized period by period).
func IjarahLeasePostings(id string, p IjarahParams) []core.Posting {
	return []core.Posting{
		{Source: AssetAccount(id), Destination: InUseAccount(p.Client), Asset: p.AssetCode, Amount: 1},
	}
}

// IjarahPayRentPostings — LEASED → LEASED (or → COMPLETED when last).
// P1 cash rent in; P2 recognizes the rent as income (FAS 32 simplified);
// P3 depreciates this period's share of the book value. On the last period
// the unit is returned to the lessor.
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

// IjarahPenaltyPostings — late penalty. Hard rule: charity-only destination
// (AAOIFI SS 3); a penalty can never become bank income.
func IjarahPenaltyPostings(id string, p IjarahParams, amount int64, destination string) ([]core.Posting, error) {
	if !strings.HasPrefix(destination, CharityPrefix) {
		return nil, &Error{Code: ErrShariaViolation,
			Message: "late penalty destination must be a @charity: account", StandardRef: RefSS3, ContractID: id}
	}
	return []core.Posting{
		{Source: p.Client, Destination: destination, Asset: p.Cost.Asset, Amount: amount},
	}, nil
}

// IjarahCancelPostings — PROMISE → CANCELLED (no postings) or
// ACQUIRED → CANCELLED (bank keeps the asset, its ownership risk).
func IjarahCancelPostings(id string, p IjarahParams, fromState string) []core.Posting {
	if fromState != StateAcquired {
		return nil
	}
	return []core.Posting{
		{Source: AssetAccount(id), Destination: UnsoldInventory, Asset: p.AssetCode, Amount: 1},
		{Source: AssetAccount(id), Destination: core.WORLD, Asset: p.Cost.Asset, Amount: p.Cost.Amount},
	}
}

// ---------------------------------------------------------------------------
// ContractKind implementation
// ---------------------------------------------------------------------------

type ijarahKind struct{}

func init() { register(ijarahKind{}) }

func (ijarahKind) Type() string { return TypeIjarah }

func (ijarahKind) DecodeParams(raw json.RawMessage) (Params, error) {
	var p IjarahParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, newError(ErrInvalidParams, "invalid ijarah params: "+err.Error())
	}
	p.applyDefaults()
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

// ShariaGate: leasing a non-owned asset is an SS-9 violation, checked before
// the FSM (even from a wrong state), mirroring murabaha's qabd gate.
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
			if ev.Input.Seq > 0 {
				return &Error{Code: ErrNotFound, Message: "installment does not exist"}
			}
			return &Error{Code: ErrPrecondition, Message: "no unpaid installment left"}
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
		// charity-gate enforced in BuildPlan via IjarahPenaltyPostings, before commit
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
				StandardRef: RefFAS32, TxID: -1, Payload: "{}"}}
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"seq": inst.Seq, "rent": inst.Amount, "depreciation_part": inst.DepreciationPart,
			"accounting": "FAS 32 simplified (v1); depreciation basis to be validated by a Sharia/accounting advisor",
		})
		return TransitionPlan{
			Postings:  IjarahPayRentPostings(c.ID, ip, *inst, last),
			Reference: fmt.Sprintf("%s:rent:%d", c.ID, inst.Seq),
			NewState:  newState, StandardRef: RefFAS32, Payload: string(payload),
			Marks:      []InstallmentMark{{Seq: inst.Seq, Status: StatusPaid}},
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

package sharia

import (
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

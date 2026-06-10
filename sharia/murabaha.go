package sharia

import (
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

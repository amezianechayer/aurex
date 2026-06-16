package payments

import "github.com/amezianechayer/corren/core"

// PaymentStore persists PSP payment records. Implemented by storage/sqlite and
// storage/postgres, embedded into storage.Store like the other module stores.
type PaymentStore interface {
	SavePayment(p Payment) error
	UpdatePayment(p Payment) error
	GetPayment(id string) (Payment, error)
	ListPayments(limit, offset int) ([]Payment, error)
}

// LedgerPort is the slice of the ledger payments needs. *ledger.Ledger satisfies
// it. Payins and payouts are committed as two-leg transactions through it, so
// they always pass the Guard, and the @psp:<name> clearing account nets to zero
// while the wallet's main account holds the real balance.
type LedgerPort interface {
	Commit(ts []core.Transaction) ([]core.Transaction, error)
}

// Registry resolves a PSP connector by name.
type Registry struct {
	conns map[string]PSP
}

func NewRegistry(psps ...PSP) *Registry {
	r := &Registry{conns: map[string]PSP{}}
	for _, p := range psps {
		if p != nil {
			r.conns[p.Name()] = p
		}
	}
	return r
}

func (r *Registry) Get(name string) (PSP, bool) {
	p, ok := r.conns[name]
	return p, ok
}

package guard

import "github.com/amezianechayer/corren/core"

// LedgerView is the minimal read surface a rule MIGHT need, passed to Evaluate
// to keep the signature stable and avoid the guard->ledger import cycle. The v1
// rules don't use it; v2 rules (e.g. balance-dependent) can.
type LedgerView interface {
	GetAccount(address string) (core.Account, error)
}

// GuardStore persists rules and events. Implemented by storage/sqlite and
// storage/postgres, embedded into storage.Store like sharia.ShariaStore.
type GuardStore interface {
	SaveRule(r Rule) error
	UpdateRule(r Rule) error
	DeleteRule(id string) error
	GetRule(id string) (Rule, error)
	ListRules() ([]Rule, error)

	AppendGuardEvent(e GuardEvent) error
	ListGuardEvents(limit, offset int) ([]GuardEvent, error)
}

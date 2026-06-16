package wallets

import "github.com/amezianechayer/corren/core"

// WalletStore persists wallet and hold metadata only — never balances, which
// always come from the ledger. Implemented by storage/sqlite and
// storage/postgres, embedded into storage.Store like sharia.ShariaStore.
type WalletStore interface {
	SaveWallet(w Wallet) error
	GetWallet(id string) (Wallet, error)
	ListWallets(limit, offset int) ([]Wallet, error)

	SaveHold(h Hold) error
	UpdateHold(h Hold) error
	GetHold(id string) (Hold, error)
	ListHolds(walletID string) ([]Hold, error)
}

// LedgerPort is the slice of the ledger the wallet service needs. *ledger.Ledger
// satisfies it. Keeping it narrow avoids a wallets->ledger import cycle and
// makes the service unit-testable with a fake. Every money movement goes
// through Commit, so it always passes the Guard.
type LedgerPort interface {
	Commit(ts []core.Transaction) ([]core.Transaction, error)
	GetAccount(address string) (core.Account, error)
	SaveMeta(targetType string, targetID string, m core.Metadata) error
}

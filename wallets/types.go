// Package wallets adds a user-facing wallet primitive on top of the corren
// ledger. A wallet never stores a balance of its own: its funds live in ledger
// accounts (@wallets:<id>:main for available, @wallets:<id>:holds for reserved),
// so every balance is the ledger's truth and every movement goes through
// ledger.Commit — which means it always passes the FaRl/Sharia Guard.
//
// Holds are Sharia-aware by construction: a hold is a real posting from the
// main account to the holds account, so the ledger's non-negative invariant
// refuses a hold on funds that do not exist — no fictitious debt.
package wallets

import (
	"crypto/rand"
	"encoding/hex"
)

// Hold lifecycle.
const (
	HoldActive   = "active"
	HoldCaptured = "captured"
	HoldVoided   = "voided"
)

// Wallet is the user-owned container. Asset is its primary denomination (e.g.
// "AED.2"); operations may target other assets, but Asset is the default.
type Wallet struct {
	ID        string `json:"id"`
	Owner     string `json:"owner"`
	Asset     string `json:"asset"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Hold records the intent and status of a reservation. The reserved funds
// themselves sit in the ledger holds account; this row tracks what they are for.
type Hold struct {
	ID          string `json:"id"`
	WalletID    string `json:"wallet_id"`
	Asset       string `json:"asset"`
	Amount      int64  `json:"amount"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Balances is the read view of a wallet: available (spendable) and held
// (reserved) amounts, per asset, both read straight from the ledger.
type Balances struct {
	WalletID  string           `json:"wallet_id"`
	Available map[string]int64 `json:"available"`
	Held      map[string]int64 `json:"held"`
}

// MainAccount is the ledger account holding a wallet's spendable funds.
func MainAccount(walletID string) string { return "@wallets:" + walletID + ":main" }

// HoldsAccount is the ledger account holding a wallet's reserved funds.
func HoldsAccount(walletID string) string { return "@wallets:" + walletID + ":holds" }

// generateID returns a prefixed random identifier (e.g. "wlt_9f3a...").
func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

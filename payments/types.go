// Package payments adds PSP connectors on top of the corren ledger + wallets,
// Sharia-native: a payin credits a wallet and a payout debits one, both as FaRl
// postings through ledger.Commit (so always through the Guard), and a payout is
// only ever initiated AFTER the wallet has been debited — you can never pay out
// money a wallet does not hold (no riba, no fictitious debt).
package payments

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// Payment direction.
const (
	DirectionPayin  = "payin"
	DirectionPayout = "payout"
)

// Payment / payout status.
const (
	StatusPending   = "pending"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

// Normalized webhook event types.
const (
	EventPaymentSucceeded = "payment.succeeded"
	EventPaymentFailed    = "payment.failed"
	EventUnhandled        = "unhandled"
)

// Payment is a persisted record of a money movement through a PSP. Both
// directions share this shape; the ledger remains the source of truth for
// balances — this row is the PSP-side audit record.
type Payment struct {
	ID         string `json:"id"`
	PSP        string `json:"psp"`
	Direction  string `json:"direction"`
	WalletID   string `json:"wallet_id"`
	Asset      string `json:"asset"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	Reference  string `json:"reference"`
	ExternalID string `json:"external_id,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`

	// ClientSecret and RedirectURL are TRANSIENT next-action credentials returned
	// only when a payin is initiated (CreatePayin): the client uses them to
	// complete the charge — ClientSecret with Stripe.js, RedirectURL for a
	// hosted-page PSP (PayTabs). They are deliberately NOT persisted (the store
	// writes only the fixed column set), so they never appear when a payment is
	// read back.
	ClientSecret string `json:"client_secret,omitempty"`
	RedirectURL  string `json:"redirect_url,omitempty"`
}

// Payout is the PSP-side result of initiating an outbound payment.
type Payout struct {
	ExternalID string `json:"external_id"`
	Status     string `json:"status"`
}

// Event is the normalized, provider-agnostic result of parsing a PSP webhook.
type Event struct {
	Type       string `json:"type"`
	PSP        string `json:"psp"`
	Reference  string `json:"reference"`
	ExternalID string `json:"external_id"`
	WalletID   string `json:"wallet_id"`
	Asset      string `json:"asset"`
	Amount     int64  `json:"amount"`
}

// PSP is the common connector interface every provider implements. HandleWebhook
// takes the raw body and the provider's signature header (the FLOW requires
// signature validation), returning a normalized Event.
type PSP interface {
	Name() string
	CreatePayment(amount int64, asset, ref string) (Payment, error)
	CreatePayout(amount int64, asset, dest, ref string) (Payout, error)
	HandleWebhook(payload []byte, signature string) (Event, error)
}

// PSPAccount is the ledger account standing in for a provider's external funds.
func PSPAccount(name string) string { return "@psp:" + name }

// currencyOf maps a corren asset ("AED.2") to a PSP currency code ("aed").
func currencyOf(asset string) string {
	if i := strings.IndexByte(asset, '.'); i >= 0 {
		asset = asset[:i]
	}
	return strings.ToLower(asset)
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

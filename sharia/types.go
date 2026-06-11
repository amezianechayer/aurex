package sharia

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"time"
)

// Contract types and states
const (
	TypeMurabaha = "murabaha"

	StatePromise   = "PROMISE"
	StateAcquired  = "ACQUIRED"
	StateSold      = "SOLD"
	StateSettled   = "SETTLED"
	StateCancelled = "CANCELLED"
)

// Installment statuses
const (
	StatusPending      = "pending"
	StatusPaid         = "paid"
	StatusOverdue      = "overdue"
	StatusSettledEarly = "settled_early"
)

// AAOIFI standard references (document-level; exact clauses to be
// confirmed by a Sharia advisor before production).
const (
	RefSS8   = "AAOIFI-SS-8"
	RefSS3   = "AAOIFI-SS-3"
	RefFAS28 = "AAOIFI-FAS-28"
)

const TemplateVersionMurabaha = "murabaha/1.0.0"

var (
	// FaRl assets: [A-Z][A-Z0-9]* with an optional dotted precision
	// suffix — the platform's historical convention is "DZD.2"/"EUR.2".
	assetRe = regexp.MustCompile(`^[A-Z][A-Z0-9]{0,7}(\.[0-9]{1,2})?$`)
	// FaRl lexer: ACCOUNT : '@' [a-z_][a-z0-9_:]*
	accountRe = regexp.MustCompile(`^@[a-z_][a-z0-9_:]*$`)
	// Contract IDs are embedded in account addresses: lowercase only, no dashes.
	contractIDRe = regexp.MustCompile(`^[a-z][a-z0-9_]{3,40}$`)
)

type Monetary struct {
	Asset  string `json:"asset"`
	Amount int64  `json:"amount"`
}

type MurabahaParams struct {
	AssetCode    string   `json:"asset_code"`
	Cost         Monetary `json:"cost"`
	Markup       Monetary `json:"markup"`
	Client       string   `json:"client"`
	Supplier     string   `json:"supplier"`
	BankTreasury string   `json:"bank_treasury"`
	Installments int      `json:"installments"`
	FirstDue     string   `json:"first_due"`
	PeriodDays   int      `json:"period_days"`
}

// Validate checks params per spec §5.1 and applies defaults
// (treasury, period). Returns ERR_INVALID_PARAMS with a per-field detail.
func (p *MurabahaParams) Validate() error {
	if p.BankTreasury == "" {
		p.BankTreasury = "@bank:treasury"
	}
	if p.PeriodDays == 0 {
		p.PeriodDays = 30
	}

	switch {
	case p.Cost.Amount <= 0:
		return newError(ErrInvalidParams, "cost.amount must be > 0")
	case p.Markup.Amount < 0:
		return newError(ErrInvalidParams, "markup.amount must be >= 0")
	case !assetRe.MatchString(p.Cost.Asset):
		return newError(ErrInvalidParams, "cost.asset is not a valid asset code")
	case p.Cost.Asset != p.Markup.Asset:
		return newError(ErrInvalidParams, "cost and markup asset must match")
	case !assetRe.MatchString(p.AssetCode):
		return newError(ErrInvalidParams, "asset_code is not a valid asset code")
	case p.AssetCode == p.Cost.Asset:
		return newError(ErrInvalidParams, "asset_code must differ from the monetary asset")
	case p.Installments < 1:
		return newError(ErrInvalidParams, "installments must be >= 1")
	case !accountRe.MatchString(p.Client):
		return newError(ErrInvalidParams, "client is not a valid account address")
	case !accountRe.MatchString(p.Supplier):
		return newError(ErrInvalidParams, "supplier is not a valid account address")
	case !accountRe.MatchString(p.BankTreasury):
		return newError(ErrInvalidParams, "bank_treasury is not a valid account address")
	case p.PeriodDays < 1:
		return newError(ErrInvalidParams, "period_days must be >= 1")
	}

	if _, err := time.Parse(time.RFC3339, p.FirstDue); err != nil {
		return newError(ErrInvalidParams, "first_due is not a valid RFC3339 date")
	}
	return nil
}

type Contract struct {
	ID              string         `json:"id"`
	Type            string         `json:"type"`
	State           string         `json:"state"`
	Params          MurabahaParams `json:"params"`
	TemplateVersion string         `json:"template_version"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
}

type Installment struct {
	Seq           int    `json:"seq"`
	DueDate       string `json:"due_date"`
	Amount        int64  `json:"amount"`
	PrincipalPart int64  `json:"principal_part"`
	ProfitPart    int64  `json:"profit_part"`
	Status        string `json:"status"`
	PaidTxID      int64  `json:"paid_tx_id"`
	PaidAt        string `json:"paid_at,omitempty"`
}

type DueItem struct {
	ContractID string `json:"contract_id"`
	Seq        int    `json:"seq"`
	DueDate    string `json:"due_date"`
	Amount     int64  `json:"amount"`
}

func ContractIDIsValid(id string) bool {
	return contractIDRe.MatchString(id)
}

// GenerateContractID returns ids like "mur20260610a7x4" — lowercase
// alphanumeric only, compatible with ledger account addresses (§4.1).
func GenerateContractID() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 4)
	rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return fmt.Sprintf("mur%s%s", time.Now().UTC().Format("20060102"), string(b))
}

// Account addresses per contract (spec §4.1)
func InventoryAccount(id string) string   { return "@contracts:" + id + ":inventory" }
func ReceivableAccount(id string) string  { return "@contracts:" + id + ":receivable" }
func DeferredAccount(id string) string    { return "@contracts:" + id + ":deferred" }
func CounterpartAccount(id string) string { return "@contracts:" + id + ":counterpart" }

const (
	IncomeAccount       = "@bank:income:murabaha"
	CharityPrefix       = "@charity:"
	UnsoldInventory     = "@bank:inventory:unsold"
	DefaultCharityPool  = "@charity:pool"
	DefaultBankTreasury = "@bank:treasury"
)

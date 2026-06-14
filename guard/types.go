package guard

import "encoding/json"

// Rule kinds (v1)
const (
	KindAmountCap     = "amount_cap"
	KindAccountList   = "account_list"
	KindAssetRestrict = "asset_restrict"

	ActionDeny    = "deny"
	ActionMonitor = "monitor"
)

// Rule is a declarative guard rule, stored as JSON params per kind.
type Rule struct {
	ID          string          `json:"id"`
	Kind        string          `json:"kind"`
	Params      json.RawMessage `json:"params"`
	Action      string          `json:"action"`       // deny | monitor
	Reason      string          `json:"reason"`       // always required
	StandardRef string          `json:"standard_ref"` // required when action == deny
	Enabled     bool            `json:"enabled"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// GuardEvent records one rule firing (deny or monitor).
type GuardEvent struct {
	ID          int64  `json:"id"`
	RuleID      string `json:"rule_id"`
	Action      string `json:"action"`
	Reason      string `json:"reason"`
	StandardRef string `json:"standard_ref"`
	TxReference string `json:"tx_reference"`
	Payload     string `json:"payload"`
	CreatedAt   string `json:"created_at"`
}

// Per-kind param structs.
type AmountCapParams struct {
	Scope string `json:"scope"`
	Asset string `json:"asset"`
	Max   int64  `json:"max"`
	Basis string `json:"basis"` // "posting" | "net_outflow"
}

type AccountListParams struct {
	Mode     string   `json:"mode"` // "block" | "allow"
	Side     string   `json:"side"` // "source" | "destination" | "either"
	Patterns []string `json:"patterns"`
}

type AssetRestrictParams struct {
	Scope  string   `json:"scope"`
	Mode   string   `json:"mode"` // "only" | "never"
	Assets []string `json:"assets"`
}

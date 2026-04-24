package core

type ContractType string

const (
	ContractMurabaha   ContractType = "MURABAHA"
	ContractMudarabah  ContractType = "MUDARABAH"
	ContractMusharakah ContractType = "MUSHARAKAH"
	ContractIjara      ContractType = "IJARA"
	ContractSukuk      ContractType = "SUKUK"
	ContractQardHassan ContractType = "QARD_HASSAN"
)

type ContractStatus string

const (
	ContractPending   ContractStatus = "PENDING"
	ContractActive    ContractStatus = "ACTIVE"
	ContractCompleted ContractStatus = "COMPLETED"
	ContractCancelled ContractStatus = "CANCELLED"
)

var AaoifiFAS = map[ContractType]string{
	ContractMurabaha:   "FAS-2",
	ContractMudarabah:  "FAS-3",
	ContractMusharakah: "FAS-4",
	ContractIjara:      "FAS-32",
	ContractSukuk:      "FAS-33",
	ContractQardHassan: "FAS-26",
}

// RequiredTerms defines mandatory fields in contract.Terms per contract type.
// Missing fields = gharar (contractual ambiguity).
var RequiredTerms = map[ContractType][]string{
	ContractMurabaha:   {"markup", "asset_description"},
	ContractMudarabah:  {"profit_ratio", "capital"},
	ContractMusharakah: {"profit_ratio", "loss_ratio"},
	ContractIjara:      {"rent_amount", "asset_description", "duration"},
	ContractSukuk:      {"underlying_asset", "face_value"},
	ContractQardHassan: {},
}

type ShariaContract struct {
	ID        string                 `json:"id"`
	Type      ContractType           `json:"type"`
	Status    ContractStatus         `json:"status"`
	Ledger    string                 `json:"ledger"`
	Parties   map[string]string      `json:"parties"`
	Terms     map[string]interface{} `json:"terms"`
	AaoifiFAS string                 `json:"aaoifi_fas"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type ShariaCertificate struct {
	ID         string `json:"id"`
	ContractID string `json:"contract_id,omitempty"`
	TxID       int64  `json:"txid,omitempty"`
	Constraint string `json:"constraint"`
	Result     string `json:"result"`
	IssuedAt   string `json:"issued_at"`
	Authority  string `json:"authority"`
}

type AssetCategory string

const (
	AssetCommodity  AssetCategory = "COMMODITY"
	AssetCurrency   AssetCategory = "CURRENCY"
	AssetEquity     AssetCategory = "EQUITY"
	AssetRealEstate AssetCategory = "REAL_ESTATE"
	AssetDebt       AssetCategory = "DEBT"
)

type AssetEntry struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Precision   int               `json:"precision"`
	Category    AssetCategory     `json:"category"`
	AaoifiClass string            `json:"aaoifi_class"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   string            `json:"created_at"`
}

type APIKey struct {
	ID        int64  `json:"id"`
	KeyHash   string `json:"-"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

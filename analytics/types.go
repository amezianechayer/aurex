package analytics

type AssetVolume struct {
	Asset string `json:"asset"`
	Total int64  `json:"total"`
}

type AccountVolume struct {
	Account string `json:"account"`
	Asset   string `json:"asset"`
	Volume  int64  `json:"volume"`
}

type Overview struct {
	Transactions  int64           `json:"transactions"`
	Accounts      int64           `json:"accounts"`
	VolumeByAsset []AssetVolume   `json:"volume_by_asset"`
	TopAccounts   []AccountVolume `json:"top_accounts"`
}

type FlowEdge struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Asset       string `json:"asset"`
	TimeBucket  string `json:"time_bucket"`
	Amount      int64  `json:"amount"`
	Count       int64  `json:"count"`
}

type Rollup struct {
	ContractsByState    map[string]int64 `json:"contracts_by_state"`
	GuardEventsByAction map[string]int64 `json:"guard_events_by_action"`
}

type TimeBucket struct {
	TimeBucket string `json:"time_bucket"`
	In         int64  `json:"in"`
	Out        int64  `json:"out"`
}

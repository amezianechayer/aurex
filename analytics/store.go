package analytics

// AnalyticsStore is the read-only aggregation surface for FaRl Lens.
// Implemented by storage/sqlite and storage/postgres, embedded into
// storage.Store like sharia.ShariaStore and guard.GuardStore.
type AnalyticsStore interface {
	LensOverview() (Overview, error)
	LensFlows(limit int) ([]FlowEdge, error)
	LensRollup() (Rollup, error)
	LensTimeSeries(account, asset string) ([]TimeBucket, error)
}

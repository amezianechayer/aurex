package sqlite

import (
	"github.com/amezianechayer/corren/analytics"
)

// LensOverview returns transaction/account counts, volume per asset, and the
// top-10 accounts by total flow (source + destination combined).
func (s *SQLiteStore) LensOverview() (analytics.Overview, error) {
	ov := analytics.Overview{
		VolumeByAsset: []analytics.AssetVolume{},
		TopAccounts:   []analytics.AccountVolume{},
	}

	txCount, err := s.CountTransactions()
	if err != nil {
		return ov, err
	}
	ov.Transactions = txCount

	accCount, err := s.CountAccounts()
	if err != nil {
		return ov, err
	}
	ov.Accounts = accCount

	// Volume by asset: total amount moved per asset across all postings.
	rows, err := s.db.Query(`SELECT asset, SUM(amount) FROM postings GROUP BY asset ORDER BY asset`)
	if err != nil {
		return ov, err
	}
	defer rows.Close()
	for rows.Next() {
		var av analytics.AssetVolume
		if err := rows.Scan(&av.Asset, &av.Total); err != nil {
			return ov, err
		}
		ov.VolumeByAsset = append(ov.VolumeByAsset, av)
	}
	if err := rows.Err(); err != nil {
		return ov, err
	}

	// Top accounts: union source and destination, sum by (account, asset).
	topRows, err := s.db.Query(`
		SELECT account, asset, SUM(amount) AS vol
		FROM (
			SELECT source AS account, asset, amount FROM postings
			UNION ALL
			SELECT destination AS account, asset, amount FROM postings
		)
		GROUP BY account, asset
		ORDER BY vol DESC
		LIMIT 10
	`)
	if err != nil {
		return ov, err
	}
	defer topRows.Close()
	for topRows.Next() {
		var av analytics.AccountVolume
		if err := topRows.Scan(&av.Account, &av.Asset, &av.Volume); err != nil {
			return ov, err
		}
		ov.TopAccounts = append(ov.TopAccounts, av)
	}
	if err := topRows.Err(); err != nil {
		return ov, err
	}

	return ov, nil
}

// LensFlows returns aggregated flow edges (source→destination per asset per day).
// limit <= 0 defaults to 100.
func (s *SQLiteStore) LensFlows(limit int) ([]analytics.FlowEdge, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(`
		SELECT p.source, p.destination, p.asset,
		       substr(t.timestamp, 1, 10) AS tb,
		       SUM(p.amount) AS amount,
		       COUNT(*) AS cnt
		FROM postings p
		JOIN transactions t ON p.txid = t.id
		GROUP BY p.source, p.destination, p.asset, tb
		ORDER BY amount DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	edges := []analytics.FlowEdge{}
	for rows.Next() {
		var e analytics.FlowEdge
		if err := rows.Scan(&e.Source, &e.Destination, &e.Asset, &e.TimeBucket, &e.Amount, &e.Count); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return edges, nil
}

// LensRollup returns contract counts by state and guard-event counts by action.
// Both maps are always non-nil (even on an empty ledger).
func (s *SQLiteStore) LensRollup() (analytics.Rollup, error) {
	r := analytics.Rollup{
		ContractsByState:    map[string]int64{},
		GuardEventsByAction: map[string]int64{},
	}

	cRows, err := s.db.Query(`SELECT state, COUNT(*) FROM sharia_contracts GROUP BY state`)
	if err != nil {
		return r, err
	}
	defer cRows.Close()
	for cRows.Next() {
		var state string
		var cnt int64
		if err := cRows.Scan(&state, &cnt); err != nil {
			return r, err
		}
		r.ContractsByState[state] = cnt
	}
	if err := cRows.Err(); err != nil {
		return r, err
	}

	gRows, err := s.db.Query(`SELECT action, COUNT(*) FROM guard_events GROUP BY action`)
	if err != nil {
		return r, err
	}
	defer gRows.Close()
	for gRows.Next() {
		var action string
		var cnt int64
		if err := gRows.Scan(&action, &cnt); err != nil {
			return r, err
		}
		r.GuardEventsByAction[action] = cnt
	}
	if err := gRows.Err(); err != nil {
		return r, err
	}

	return r, nil
}

// LensTimeSeries returns per-day in/out flow for a given account and asset.
func (s *SQLiteStore) LensTimeSeries(account, asset string) ([]analytics.TimeBucket, error) {
	rows, err := s.db.Query(`
		SELECT substr(t.timestamp, 1, 10) AS tb,
		       SUM(CASE WHEN p.destination = ? THEN p.amount ELSE 0 END) AS inn,
		       SUM(CASE WHEN p.source      = ? THEN p.amount ELSE 0 END) AS outt
		FROM postings p
		JOIN transactions t ON p.txid = t.id
		WHERE (p.source = ? OR p.destination = ?) AND p.asset = ?
		GROUP BY tb
		ORDER BY tb ASC
	`, account, account, account, account, asset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buckets := []analytics.TimeBucket{}
	for rows.Next() {
		var b analytics.TimeBucket
		if err := rows.Scan(&b.TimeBucket, &b.In, &b.Out); err != nil {
			return nil, err
		}
		buckets = append(buckets, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buckets, nil
}

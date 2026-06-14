package postgres

import (
	"context"
	"fmt"

	"github.com/amezianechayer/corren/analytics"
)

// LensOverview returns transaction/account counts, volume per asset, and the
// top-10 accounts by total flow (source + destination combined).
func (s *PGStore) LensOverview() (analytics.Overview, error) {
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
	volSQL := fmt.Sprintf(
		`SELECT asset, SUM(amount) FROM %s GROUP BY asset ORDER BY asset`,
		s.table("postings"),
	)
	rows, err := s.Conn().Query(context.Background(), volSQL)
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
	// Postgres requires an alias on the subquery ("sub").
	topSQL := fmt.Sprintf(`
		SELECT account, asset, SUM(amount) AS vol
		FROM (
			SELECT source AS account, asset, amount FROM %s
			UNION ALL
			SELECT destination AS account, asset, amount FROM %s
		) sub
		GROUP BY account, asset
		ORDER BY vol DESC
		LIMIT 10
	`, s.table("postings"), s.table("postings"))
	topRows, err := s.Conn().Query(context.Background(), topSQL)
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
func (s *PGStore) LensFlows(limit int) ([]analytics.FlowEdge, error) {
	if limit <= 0 {
		limit = 100
	}

	flowSQL := fmt.Sprintf(`
		SELECT p.source, p.destination, p.asset,
		       substr(t.timestamp, 1, 10) AS tb,
		       SUM(p.amount) AS amount,
		       COUNT(*) AS cnt
		FROM %s p
		JOIN %s t ON p.txid = t.id
		GROUP BY p.source, p.destination, p.asset, tb
		ORDER BY amount DESC
		LIMIT $1
	`, s.table("postings"), s.table("transactions"))

	rows, err := s.Conn().Query(context.Background(), flowSQL, limit)
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
func (s *PGStore) LensRollup() (analytics.Rollup, error) {
	r := analytics.Rollup{
		ContractsByState:    map[string]int64{},
		GuardEventsByAction: map[string]int64{},
	}

	cSQL := fmt.Sprintf(
		`SELECT state, COUNT(*) FROM %s GROUP BY state`,
		s.table("sharia_contracts"),
	)
	cRows, err := s.Conn().Query(context.Background(), cSQL)
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

	gSQL := fmt.Sprintf(
		`SELECT action, COUNT(*) FROM %s GROUP BY action`,
		s.table("guard_events"),
	)
	gRows, err := s.Conn().Query(context.Background(), gSQL)
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
func (s *PGStore) LensTimeSeries(account, asset string) ([]analytics.TimeBucket, error) {
	tsSQL := fmt.Sprintf(`
		SELECT substr(t.timestamp, 1, 10) AS tb,
		       SUM(CASE WHEN p.destination = $1 THEN p.amount ELSE 0 END) AS inn,
		       SUM(CASE WHEN p.source      = $2 THEN p.amount ELSE 0 END) AS outt
		FROM %s p
		JOIN %s t ON p.txid = t.id
		WHERE (p.source = $3 OR p.destination = $4) AND p.asset = $5
		GROUP BY tb
		ORDER BY tb ASC
	`, s.table("postings"), s.table("transactions"))

	rows, err := s.Conn().Query(context.Background(), tsSQL, account, account, account, account, asset)
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

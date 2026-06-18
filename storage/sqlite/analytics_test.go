package sqlite

import (
	"os"
	"path"
	"testing"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/guard"
	"github.com/spf13/viper"
)

// seedLens builds a known ledger:
//
//	tx0 2026-07-01: @world -> @client:a [DZD.2 1000]
//	tx1 2026-07-02: @client:a -> @bank:treasury [DZD.2 400]
//	tx2 2026-07-02: @world -> @client:a [EUR.2 50]
//
// plus one PROMISE contract and one deny guard event (for the rollup).
func seedLens(t *testing.T, s *SQLiteStore) {
	t.Helper()
	txs := []core.Transaction{
		{ID: 0, Timestamp: "2026-07-01T10:00:00Z", Postings: []core.Posting{
			{Source: "@world", Destination: "@client:a", Amount: 1000, Asset: "DZD.2"}}},
		{ID: 1, Timestamp: "2026-07-02T10:00:00Z", Postings: []core.Posting{
			{Source: "@client:a", Destination: "@bank:treasury", Amount: 400, Asset: "DZD.2"}}},
		{ID: 2, Timestamp: "2026-07-02T11:00:00Z", Postings: []core.Posting{
			{Source: "@world", Destination: "@client:a", Amount: 50, Asset: "EUR.2"}}},
	}
	if err := s.SaveTransactions(txs); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`INSERT INTO sharia_contracts (id, type, state, params, template_version, created_at, updated_at) VALUES (?,?,?,?,?,?,?)`,
		"c1", "murabaha", "PROMISE", "{}", "x", "2026-07-01T00:00:00Z", "2026-07-01T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendGuardEvent(guard.GuardEvent{RuleID: "r1", Action: guard.ActionDeny,
		Reason: "x", StandardRef: "P", CreatedAt: "2026-07-02T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}
}

func withLensStore(t *testing.T, f func(s *SQLiteStore)) {
	t.Helper()
	// Remove any leftover DB from a previous run so seeded and empty tests
	// don't contaminate each other.
	os.Remove(path.Join(viper.GetString("storage.dir"), viper.GetString("storage.sqlite.db_name")+"_lenstest.db"))
	s, err := NewStore("lenstest")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	f(s)
}

func TestLensOverview(t *testing.T) {
	withLensStore(t, func(s *SQLiteStore) {
		seedLens(t, s)
		ov, err := s.LensOverview()
		if err != nil {
			t.Fatal(err)
		}
		if ov.Transactions != 3 {
			t.Fatalf("transactions: want 3 got %d", ov.Transactions)
		}
		vol := map[string]int64{}
		for _, v := range ov.VolumeByAsset {
			vol[v.Asset] = v.Total
		}
		if vol["DZD.2"] != 1400 || vol["EUR.2"] != 50 {
			t.Fatalf("volume_by_asset wrong: %+v", ov.VolumeByAsset)
		}
		if len(ov.TopAccounts) == 0 {
			t.Fatal("expected top accounts")
		}
		if ov.TopAccounts[0].Account != "@client:a" || ov.TopAccounts[0].Asset != "DZD.2" || ov.TopAccounts[0].Volume != 1400 {
			t.Fatalf("top account wrong: %+v", ov.TopAccounts[0])
		}
	})
}

func TestLensFlows(t *testing.T) {
	withLensStore(t, func(s *SQLiteStore) {
		seedLens(t, s)
		edges, err := s.LensFlows(100)
		if err != nil {
			t.Fatal(err)
		}
		var amount, count int64 = -1, -1
		for _, e := range edges {
			if e.Source == "@world" && e.Destination == "@client:a" && e.Asset == "DZD.2" && e.TimeBucket == "2026-07-01" {
				amount, count = e.Amount, e.Count
			}
		}
		if amount != 1000 || count != 1 {
			t.Fatalf("expected world->client:a DZD.2 2026-07-01 amount=1000 count=1, edges=%+v", edges)
		}
	})
}

func TestLensRollup(t *testing.T) {
	withLensStore(t, func(s *SQLiteStore) {
		seedLens(t, s)
		r, err := s.LensRollup()
		if err != nil {
			t.Fatal(err)
		}
		if r.ContractsByState["PROMISE"] != 1 {
			t.Fatalf("contracts_by_state: %+v", r.ContractsByState)
		}
		if r.GuardEventsByAction["deny"] != 1 {
			t.Fatalf("guard_events_by_action: %+v", r.GuardEventsByAction)
		}
	})
}

func TestLensTimeSeries(t *testing.T) {
	withLensStore(t, func(s *SQLiteStore) {
		seedLens(t, s)
		ts, err := s.LensTimeSeries("@client:a", "DZD.2")
		if err != nil {
			t.Fatal(err)
		}
		got := map[string][2]int64{}
		for _, b := range ts {
			got[b.TimeBucket] = [2]int64{b.In, b.Out}
		}
		if got["2026-07-01"] != [2]int64{1000, 0} {
			t.Fatalf("2026-07-01 wrong: %+v", got["2026-07-01"])
		}
		if got["2026-07-02"] != [2]int64{0, 400} {
			t.Fatalf("2026-07-02 wrong: %+v", got["2026-07-02"])
		}
	})
}

func TestLensEmptyLedger(t *testing.T) {
	withLensStore(t, func(s *SQLiteStore) {
		// fresh store, no seed; aggregations must not error and maps must be non-nil
		ov, err := s.LensOverview()
		if err != nil {
			t.Fatal(err)
		}
		_ = ov
		r, err := s.LensRollup()
		if err != nil {
			t.Fatal(err)
		}
		if r.ContractsByState == nil || r.GuardEventsByAction == nil {
			t.Fatal("rollup maps must be non-nil on empty ledger")
		}
		edges, err := s.LensFlows(100)
		if err != nil {
			t.Fatal(err)
		}
		_ = edges
		ts, err := s.LensTimeSeries("@nobody", "DZD.2")
		if err != nil {
			t.Fatal(err)
		}
		_ = ts
	})
}

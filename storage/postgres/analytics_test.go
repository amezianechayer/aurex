package postgres

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/sharia"
	"github.com/spf13/viper"
)

// withPGLensStore is the skip-gated helper for Lens integration tests.
// Mirrors withPGStore from sharia_test.go, using a unique schema prefix "pglens".
func withPGLensStore(t *testing.T, f func(s *PGStore)) {
	t.Helper()
	conn := os.Getenv("CORREN_TEST_PG_CONN")
	if conn == "" {
		t.Skip("CORREN_TEST_PG_CONN not set")
	}
	viper.Set("storage.postgres.conn_string", conn)

	name := fmt.Sprintf("pglens%d", time.Now().UnixNano())
	s, err := NewStore(name)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Initialize(); err != nil {
		t.Fatal(err)
	}
	f(s)
}

// seedLensPG populates a known ledger identical to the sqlite seedLens:
//
//	tx0 2026-07-01: @world -> @client:a [DZD.2 1000]
//	tx1 2026-07-02: @client:a -> @bank:treasury [DZD.2 400]
//	tx2 2026-07-02: @world -> @client:a [EUR.2 50]
//
// plus one PROMISE contract and one deny guard event (for the rollup).
func seedLensPG(t *testing.T, s *PGStore) {
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
	if err := s.SaveContract(sharia.Contract{
		ID: "c1", Type: sharia.TypeMurabaha,
		State: sharia.StatePromise, Params: []byte("{}"), TemplateVersion: "x",
		CreatedAt: "2026-07-01T00:00:00Z", UpdatedAt: "2026-07-01T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendGuardEvent(guard.GuardEvent{
		RuleID: "r1", Action: guard.ActionDeny,
		Reason: "x", StandardRef: "P", CreatedAt: "2026-07-02T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestPGLensOverview(t *testing.T) {
	withPGLensStore(t, func(s *PGStore) {
		seedLensPG(t, s)
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

func TestPGLensFlows(t *testing.T) {
	withPGLensStore(t, func(s *PGStore) {
		seedLensPG(t, s)
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

func TestPGLensRollup(t *testing.T) {
	withPGLensStore(t, func(s *PGStore) {
		seedLensPG(t, s)
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

func TestPGLensTimeSeries(t *testing.T) {
	withPGLensStore(t, func(s *PGStore) {
		seedLensPG(t, s)
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

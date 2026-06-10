package postgres

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/amezianechayer/corren/sharia"
	"github.com/spf13/viper"
)

// Integration tests against a real Postgres. Skipped unless
// CORREN_TEST_PG_CONN is set, e.g.:
//   CORREN_TEST_PG_CONN="postgresql://ledger:ledger@localhost:5433/ledger" go test ./storage/postgres/
func withPGStore(t *testing.T, f func(s *PGStore)) {
	t.Helper()
	conn := os.Getenv("CORREN_TEST_PG_CONN")
	if conn == "" {
		t.Skip("CORREN_TEST_PG_CONN not set")
	}
	viper.Set("storage.postgres.conn_string", conn)

	name := fmt.Sprintf("pgsh%d", time.Now().UnixNano())
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

func pgTestContract(id string) sharia.Contract {
	return sharia.Contract{
		ID:    id,
		Type:  sharia.TypeMurabaha,
		State: sharia.StatePromise,
		Params: sharia.MurabahaParams{
			AssetCode:    "VHCL42A",
			Cost:         sharia.Monetary{Asset: "SAR2", Amount: 10000000},
			Markup:       sharia.Monetary{Asset: "SAR2", Amount: 1000000},
			Client:       "@client:pg",
			Supplier:     "@supplier:pg",
			BankTreasury: "@bank:treasury",
			Installments: 24,
			FirstDue:     "2026-07-01T00:00:00Z",
			PeriodDays:   30,
		},
		TemplateVersion: sharia.TemplateVersionMurabaha,
		CreatedAt:       "2026-06-10T00:00:00Z",
		UpdatedAt:       "2026-06-10T00:00:00Z",
	}
}

func TestPGContractRoundTrip(t *testing.T) {
	withPGStore(t, func(s *PGStore) {
		if err := s.SaveContract(pgTestContract("mur_pg_rt")); err != nil {
			t.Fatal(err)
		}
		got, err := s.GetContract("mur_pg_rt")
		if err != nil || got.Params.Cost.Amount != 10000000 || got.State != sharia.StatePromise {
			t.Fatalf("got %+v err %v", got, err)
		}
		if err := s.SaveContract(pgTestContract("mur_pg_rt")); err == nil {
			t.Fatal("duplicate id must fail")
		}
		if err := s.UpdateContractState("mur_pg_rt", sharia.StateAcquired, "2026-06-11T00:00:00Z"); err != nil {
			t.Fatal(err)
		}
		got, _ = s.GetContract("mur_pg_rt")
		if got.State != sharia.StateAcquired {
			t.Fatalf("state not updated: %+v", got)
		}
		list, err := s.ListContracts(10, 0)
		if err != nil || len(list) != 1 {
			t.Fatalf("list: %v %v", list, err)
		}
	})
}

func TestPGScheduleRoundTrip(t *testing.T) {
	withPGStore(t, func(s *PGStore) {
		if err := s.SaveContract(pgTestContract("mur_pg_sched")); err != nil {
			t.Fatal(err)
		}
		items, _ := sharia.BuildSchedule(10000000, 1000000, 24, "2026-07-01T00:00:00Z", 30)
		if err := s.SaveSchedule("mur_pg_sched", items); err != nil {
			t.Fatal(err)
		}
		got, err := s.GetSchedule("mur_pg_sched")
		if err != nil || len(got) != 24 || got[0].Amount != 458333 || got[23].Amount != 458341 {
			t.Fatalf("schedule mismatch: %d err %v", len(got), err)
		}
		if err := s.MarkInstallment("mur_pg_sched", 1, sharia.StatusPaid, 42, "2026-07-01T10:00:00Z"); err != nil {
			t.Fatal(err)
		}
		got, _ = s.GetSchedule("mur_pg_sched")
		if got[0].Status != sharia.StatusPaid || got[0].PaidTxID != 42 {
			t.Fatalf("not marked: %+v", got[0])
		}

		// FindDue with cutoff math: due 2026-07-01, grace 7 → due on 2026-07-09
		due, err := s.FindDue("2026-08-15T00:00:00Z", 7)
		if err != nil {
			t.Fatal(err)
		}
		if len(due) != 1 || due[0].Seq != 2 {
			t.Fatalf("expected seq 2 due, got %+v", due)
		}
	})
}

func TestPGAuditRoundTrip(t *testing.T) {
	withPGStore(t, func(s *PGStore) {
		seq, hash, err := s.LastAuditHash("mur_pg_audit")
		if err != nil || seq != -1 || hash != "" {
			t.Fatalf("expected empty chain, got (%d,%q,%v)", seq, hash, err)
		}
		prev := sharia.GenesisHash("mur_pg_audit")
		for i := 0; i < 3; i++ {
			e := sharia.AuditEvent{
				ContractID: "mur_pg_audit", Seq: i, Event: sharia.EventTransition,
				Transition: "acquire", Decision: sharia.DecisionAllowed,
				TxID: int64(i), Payload: "{}", PrevHash: prev,
				CreatedAt: "2026-06-10T00:00:00Z",
			}
			e.Hash = sharia.ComputeAuditHash(prev, e)
			prev = e.Hash
			if err := s.AppendAudit(e); err != nil {
				t.Fatal(err)
			}
		}
		events, err := s.GetAudit("mur_pg_audit")
		if err != nil || len(events) != 3 {
			t.Fatalf("events %d err %v", len(events), err)
		}
		if !sharia.VerifyChain("mur_pg_audit", events) {
			t.Fatal("persisted chain must verify")
		}
		seq, hash, _ = s.LastAuditHash("mur_pg_audit")
		if seq != 2 || hash != events[2].Hash {
			t.Fatalf("last hash mismatch (%d %q)", seq, hash)
		}
	})
}

package sqlite

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

	"github.com/amezianechayer/corren/sharia"
	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	viper.Set("storage.dir", os.TempDir())
	viper.Set("storage.sqlite.db_name", "sharia_store")
	os.Remove(path.Join(os.TempDir(), "sharia_store_shtest.db"))
	os.Exit(m.Run())
}

func withStore(t *testing.T, f func(s *SQLiteStore)) {
	t.Helper()
	s, err := NewStore("shtest")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	f(s)
}

func testContract(id string) sharia.Contract {
	return sharia.Contract{
		ID:    id,
		Type:  sharia.TypeMurabaha,
		State: sharia.StatePromise,
		Params: sharia.MurabahaParams{
			AssetCode:    "VHCL42A",
			Cost:         sharia.Monetary{Asset: "SAR2", Amount: 10000000},
			Markup:       sharia.Monetary{Asset: "SAR2", Amount: 1000000},
			Client:       "@client:ameziane",
			Supplier:     "@supplier:toyota",
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

func TestContractRoundTrip(t *testing.T) {
	withStore(t, func(s *SQLiteStore) {
		c := testContract("mur_roundtrip")
		if err := s.SaveContract(c); err != nil {
			t.Fatal(err)
		}

		got, err := s.GetContract("mur_roundtrip")
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != c.ID || got.State != sharia.StatePromise || got.Type != sharia.TypeMurabaha {
			t.Fatalf("contract mismatch: %+v", got)
		}
		if got.Params.Cost.Amount != 10000000 || got.Params.Client != "@client:ameziane" {
			t.Fatalf("params mismatch: %+v", got.Params)
		}

		// duplicate id must fail
		if err := s.SaveContract(c); err == nil {
			t.Fatal("expected duplicate contract id to fail")
		}

		if err := s.UpdateContractState("mur_roundtrip", sharia.StateAcquired, "2026-06-11T00:00:00Z"); err != nil {
			t.Fatal(err)
		}
		got, _ = s.GetContract("mur_roundtrip")
		if got.State != sharia.StateAcquired || got.UpdatedAt != "2026-06-11T00:00:00Z" {
			t.Fatalf("state not updated: %+v", got)
		}

		list, err := s.ListContracts(10, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(list) == 0 {
			t.Fatal("expected at least one contract in list")
		}
	})
}

func TestGetContractNotFound(t *testing.T) {
	withStore(t, func(s *SQLiteStore) {
		_, err := s.GetContract("does_not_exist")
		if err == nil {
			t.Fatal("expected error for unknown contract")
		}
	})
}

func TestScheduleRoundTrip(t *testing.T) {
	withStore(t, func(s *SQLiteStore) {
		if err := s.SaveContract(testContract("mur_sched")); err != nil {
			t.Fatal(err)
		}
		items, _ := sharia.BuildSchedule(10000000, 1000000, 24, "2026-07-01T00:00:00Z", 30)
		if err := s.SaveSchedule("mur_sched", items); err != nil {
			t.Fatal(err)
		}

		got, err := s.GetSchedule("mur_sched")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 24 {
			t.Fatalf("expected 24 items, got %d", len(got))
		}
		if got[0].Seq != 1 || got[0].Amount != 458333 || got[23].Amount != 458341 {
			t.Fatalf("schedule mismatch: %+v %+v", got[0], got[23])
		}
		if got[0].Status != sharia.StatusPending || got[0].PaidTxID != -1 {
			t.Fatalf("expected pending/-1, got %+v", got[0])
		}

		if err := s.MarkInstallment("mur_sched", 1, sharia.StatusPaid, 42, "2026-07-01T10:00:00Z"); err != nil {
			t.Fatal(err)
		}
		got, _ = s.GetSchedule("mur_sched")
		if got[0].Status != sharia.StatusPaid || got[0].PaidTxID != 42 || got[0].PaidAt != "2026-07-01T10:00:00Z" {
			t.Fatalf("installment not marked: %+v", got[0])
		}
	})
}

func TestFindDue(t *testing.T) {
	withStore(t, func(s *SQLiteStore) {
		if err := s.SaveContract(testContract("mur_due")); err != nil {
			t.Fatal(err)
		}
		items, _ := sharia.BuildSchedule(3000, 300, 3, "2026-01-01T00:00:00Z", 30)
		if err := s.SaveSchedule("mur_due", items); err != nil {
			t.Fatal(err)
		}

		// now = 2026-02-05 with 7 days grace (cutoff 01-29): seq 1 (due 01-01)
		// is past grace; seq 2 (due 01-31) is within grace; seq 3 (due 03-02) not due.
		due, err := s.FindDue("2026-02-05T00:00:00Z", 7)
		if err != nil {
			t.Fatal(err)
		}
		if len(due) != 1 {
			t.Fatalf("expected 1 due item, got %d: %+v", len(due), due)
		}
		if due[0].ContractID != "mur_due" || due[0].Seq != 1 {
			t.Fatalf("unexpected due item: %+v", due[0])
		}

		// paid installments are never due
		s.MarkInstallment("mur_due", 1, sharia.StatusPaid, 1, "2026-02-04T00:00:00Z")
		due, _ = s.FindDue("2026-02-05T00:00:00Z", 7)
		if len(due) != 0 {
			t.Fatalf("expected 0 due items, got %d", len(due))
		}
	})
}

func TestAuditRoundTrip(t *testing.T) {
	withStore(t, func(s *SQLiteStore) {
		seq, hash, err := s.LastAuditHash("mur_audit")
		if err != nil {
			t.Fatal(err)
		}
		if seq != -1 || hash != "" {
			t.Fatalf("expected empty chain (-1, \"\"), got (%d, %q)", seq, hash)
		}

		prev := sharia.GenesisHash("mur_audit")
		for i := 0; i < 3; i++ {
			e := sharia.AuditEvent{
				ContractID: "mur_audit",
				Seq:        i,
				Event:      sharia.EventTransition,
				Transition: "acquire",
				Decision:   sharia.DecisionAllowed,
				TxID:       int64(i),
				Payload:    fmt.Sprintf(`{"i":%d}`, i),
				PrevHash:   prev,
				CreatedAt:  "2026-06-10T00:00:00Z",
			}
			e.Hash = sharia.ComputeAuditHash(prev, e)
			prev = e.Hash
			if err := s.AppendAudit(e); err != nil {
				t.Fatal(err)
			}
		}

		events, err := s.GetAudit("mur_audit")
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 3 {
			t.Fatalf("expected 3 events, got %d", len(events))
		}
		if !sharia.VerifyChain("mur_audit", events) {
			t.Fatal("persisted chain must verify")
		}

		seq, hash, err = s.LastAuditHash("mur_audit")
		if err != nil {
			t.Fatal(err)
		}
		if seq != 2 || hash != events[2].Hash {
			t.Fatalf("expected (2, %q), got (%d, %q)", events[2].Hash, seq, hash)
		}
	})
}

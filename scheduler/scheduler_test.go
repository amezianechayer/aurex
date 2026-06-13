package scheduler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

	"github.com/amezianechayer/corren/config"
	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/sharia"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// rawParams marshals murabaha params into the json.RawMessage CreateRequest
// now carries.
func rawParams(p sharia.MurabahaParams) json.RawMessage {
	raw, _ := json.Marshal(p)
	return raw
}

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	config.Init()
	viper.Set("storage.dir", os.TempDir())
	viper.Set("storage.sqlite.db_name", "sharia_sched")
	os.Remove(path.Join(os.TempDir(), "sharia_sched_schedtest.db"))
	os.Exit(m.Run())
}

func withLedger(t *testing.T, f func(l *ledger.Ledger)) {
	t.Helper()
	fx.New(
		fx.NopLogger,
		fx.Provide(func(lc fx.Lifecycle) (*ledger.Ledger, error) {
			return ledger.NewLedger("schedtest", lc)
		}),
		fx.Invoke(func(l *ledger.Ledger) {
			defer l.Close()
			f(l)
		}),
	)
}

func TestRunOnceMarksOverdueOnSoldContractsOnly(t *testing.T) {
	withLedger(t, func(l *ledger.Ledger) {
		e := sharia.NewEngine(l, l.Store())

		// fund and drive one contract to SOLD with past due dates
		_, err := l.Commit([]core.Transaction{{
			Postings: []core.Posting{
				{Source: core.WORLD, Destination: "@bank:treasury", Asset: "SAR2", Amount: 10000000},
			},
		}})
		if err != nil {
			t.Fatal(err)
		}

		p := sharia.MurabahaParams{
			AssetCode:    "VHCL42A",
			Cost:         sharia.Monetary{Asset: "SAR2", Amount: 10000000},
			Markup:       sharia.Monetary{Asset: "SAR2", Amount: 1000000},
			Client:       "@client:sched",
			Supplier:     "@supplier:sched",
			BankTreasury: "@bank:treasury",
			Installments: 12,
			FirstDue:     "2026-01-01T00:00:00Z",
			PeriodDays:   30,
		}
		if _, _, err := e.Create(sharia.CreateRequest{ID: "mur_sched_sold", Params: rawParams(p)}); err != nil {
			t.Fatal(err)
		}
		if _, err := e.Transition("mur_sched_sold", "acquire", sharia.TransitionInput{}); err != nil {
			t.Fatal(err)
		}
		if _, err := e.Transition("mur_sched_sold", "sell", sharia.TransitionInput{}); err != nil {
			t.Fatal(err)
		}

		// a second contract stays in PROMISE: it must NOT be touched
		p2 := p
		p2.Client = "@client:sched2"
		if _, _, err := e.Create(sharia.CreateRequest{ID: "mur_sched_promise", Params: rawParams(p2)}); err != nil {
			t.Fatal(err)
		}

		// now = 2026-02-15, grace 7 → cutoff 2026-02-08: seq 1 (01-01)
		// and seq 2 (01-31) of the SOLD contract are overdue
		marked, err := RunOnce(l, "2026-02-15T00:00:00Z", 7)
		if err != nil {
			t.Fatal(err)
		}
		if marked != 2 {
			t.Fatalf("expected 2 installments marked, got %d", marked)
		}

		items, err := l.Store().GetSchedule("mur_sched_sold")
		if err != nil {
			t.Fatal(err)
		}
		if items[0].Status != sharia.StatusOverdue || items[1].Status != sharia.StatusOverdue {
			t.Fatalf("expected seq 1+2 overdue, got %s/%s", items[0].Status, items[1].Status)
		}
		if items[2].Status != sharia.StatusPending {
			t.Fatalf("expected seq 3 pending, got %s", items[2].Status)
		}

		// PROMISE contract untouched
		items, err = l.Store().GetSchedule("mur_sched_promise")
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range items {
			if it.Status != sharia.StatusPending {
				t.Fatalf("PROMISE contract seq %d must stay pending, got %s", it.Seq, it.Status)
			}
		}

		// overdue events journalized, chain still valid
		events, err := l.Store().GetAudit("mur_sched_sold")
		if err != nil {
			t.Fatal(err)
		}
		overdueEvents := 0
		for _, ev := range events {
			if ev.Event == sharia.EventOverdue {
				overdueEvents++
			}
		}
		if overdueEvents != 2 {
			t.Fatalf("expected 2 overdue audit events, got %d", overdueEvents)
		}
		if !sharia.VerifyChain("mur_sched_sold", events) {
			t.Fatal("audit chain must stay valid")
		}

		// idempotent: a second run marks nothing new
		marked, err = RunOnce(l, "2026-02-15T00:00:00Z", 7)
		if err != nil {
			t.Fatal(err)
		}
		if marked != 0 {
			t.Fatalf("expected 0 on second run, got %d", marked)
		}
	})
}

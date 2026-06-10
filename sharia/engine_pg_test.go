package sharia_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/sharia"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Full contract lifecycle against a real Postgres — the postgres code
// paths of the ledger AND the sharia store. Skipped unless
// CORREN_TEST_PG_CONN is set.
func TestPostgresContractLifecycle(t *testing.T) {
	conn := os.Getenv("CORREN_TEST_PG_CONN")
	if conn == "" {
		t.Skip("CORREN_TEST_PG_CONN not set")
	}
	viper.Set("storage.driver", "postgres")
	viper.Set("storage.postgres.conn_string", conn)
	defer viper.Set("storage.driver", "sqlite")

	ledgerName := fmt.Sprintf("pgeng%d", time.Now().UnixNano())

	fx.New(
		fx.NopLogger,
		fx.Provide(func(lc fx.Lifecycle) (*ledger.Ledger, error) {
			return ledger.NewLedger(ledgerName, lc)
		}),
		fx.Invoke(func(l *ledger.Ledger) {
			defer l.Close()
			e := sharia.NewEngine(l, l.Store())
			const id = "mur_pg_cycle"

			fund(t, l, "@bank:treasury", "SAR2", 20000000)

			_, schedule, err := e.Create(sharia.CreateRequest{
				ID:     id,
				Params: params("@client:pg", "@supplier:pg", "@bank:treasury", 24),
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(schedule) != 24 || schedule[23].Amount != 458341 {
				t.Fatalf("bad schedule on postgres: %d items", len(schedule))
			}

			// I-1: sell before acquire → sharia violation
			_, err = e.Transition(id, "sell", sharia.TransitionInput{})
			se := shariaErr(t, err, sharia.ErrShariaViolation)
			if se.StandardRef != sharia.RefSS8 {
				t.Fatalf("expected SS-8, got %s", se.StandardRef)
			}

			mustTransition(t, e, id, "acquire")
			mustTransition(t, e, id, "sell")
			assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", 11000000)
			assertBal(t, l, "@contracts:"+id+":counterpart", "SAR2", -12000000)

			fund(t, l, "@client:pg", "SAR2", 2000000)
			var unpaid int64 = 11000000
			for seq := 1; seq <= 3; seq++ {
				if _, err := e.Transition(id, "pay_installment", sharia.TransitionInput{Seq: seq}); err != nil {
					t.Fatalf("pay %d: %v", seq, err)
				}
				unpaid -= schedule[seq-1].Amount
				assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", unpaid)
			}

			// idempotence on postgres unique reference
			_, err = e.Transition(id, "pay_installment", sharia.TransitionInput{Seq: 1})
			shariaErr(t, err, sharia.ErrDuplicate)

			// I-3 on postgres
			_, err = e.Transition(id, "late_penalty", sharia.TransitionInput{
				Seq: 4, Amount: 20000, Destination: "@bank:income:murabaha",
			})
			se = shariaErr(t, err, sharia.ErrShariaViolation)
			if se.StandardRef != sharia.RefSS3 {
				t.Fatalf("expected SS-3, got %s", se.StandardRef)
			}

			events, valid, err := e.VerifyAudit(id)
			if err != nil || !valid {
				t.Fatalf("audit chain on postgres must verify: %v", err)
			}
			if len(events) < 7 {
				t.Fatalf("expected >=7 audit events, got %d", len(events))
			}

			// scenario G equivalent: receivable cannot go negative
			_, err = l.Commit([]core.Transaction{{
				Postings: []core.Posting{{
					Source:      "@contracts:" + id + ":receivable",
					Destination: "@attacker:pg",
					Amount:      unpaid + 1, // more than its balance
					Asset:       "SAR2",
				}},
			}})
			if err == nil {
				t.Fatal("receivable must not go negative on postgres either")
			}
		}),
	)
}

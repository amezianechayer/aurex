package sharia_test

import (
	"encoding/json"
	"testing"

	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/sharia"
)

func ijarahReq(id string) sharia.CreateRequest {
	raw, _ := json.Marshal(sharia.IjarahParams{
		AssetCode: "VHCL1", Cost: sharia.Monetary{Asset: "DZD.2", Amount: 10000000},
		Rent: sharia.Monetary{Asset: "DZD.2", Amount: 500000},
		Client: "@client:anis", Supplier: "@supplier:toyota",
		BankTreasury: "@bank:treasury", Periods: 24,
		FirstDue: "2026-07-01T00:00:00Z", PeriodDays: 30,
	})
	return sharia.CreateRequest{Type: sharia.TypeIjarah, ID: id, Params: raw}
}

func TestIjarahNominalCycle(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "ijr_nominal"
		income0 := balance(t, l, "@bank:income:ijarah", "DZD.2")
		depr0 := balance(t, l, "@bank:expense:depreciation", "DZD.2")
		treas0 := balance(t, l, "@bank:treasury", "DZD.2")

		fund(t, l, "@bank:treasury", "DZD.2", 10000000)
		fund(t, l, "@client:anis", "DZD.2", 12000000)

		_, sched := mustCreateReq(t, e, ijarahReq(id))
		if len(sched) != 24 || sched[0].Amount != 500000 || sched[23].DepreciationPart != 416682 {
			t.Fatalf("bad schedule")
		}

		// lease before acquire -> SS-9 (ownership gate, pre-FSM)
		_, err := e.Transition(id, "lease", sharia.TransitionInput{})
		se := shariaErr(t, err, sharia.ErrShariaViolation)
		if se.StandardRef != sharia.RefSS9 {
			t.Fatalf("expected SS-9, got %s", se.StandardRef)
		}

		mustTransition(t, e, id, "acquire")
		assertBal(t, l, "@supplier:toyota", "DZD.2", 10000000)
		assertBal(t, l, "@contracts:"+id+":asset", "VHCL1", 1)
		assertBal(t, l, "@contracts:"+id+":asset", "DZD.2", 10000000)

		mustTransition(t, e, id, "lease")
		assertBal(t, l, "@client:anis:in_use", "VHCL1", 1)

		var res sharia.TransitionResult
		var paidDepr int64
		for seq := 1; seq <= 24; seq++ {
			res, err = e.Transition(id, "pay_rent", sharia.TransitionInput{Seq: seq})
			if err != nil {
				t.Fatalf("pay_rent %d: %v", seq, err)
			}
			paidDepr += sched[seq-1].DepreciationPart
			assertBal(t, l, "@contracts:"+id+":asset", "DZD.2", 10000000-paidDepr)
		}
		if res.NewState != sharia.StateCompleted {
			t.Fatalf("expected COMPLETED, got %s", res.NewState)
		}

		assertBal(t, l, "@contracts:"+id+":asset", "DZD.2", 0)
		assertBal(t, l, "@contracts:"+id+":asset", "VHCL1", 0)
		assertBal(t, l, "@bank:inventory:returned", "VHCL1", 1)
		assertBal(t, l, "@bank:income:ijarah", "DZD.2", income0+12000000)
		assertBal(t, l, "@bank:expense:depreciation", "DZD.2", depr0+10000000)
		assertBal(t, l, "@bank:treasury", "DZD.2", treas0+10000000-10000000+12000000)
		assertBal(t, l, "@contracts:"+id+":counterpart", "DZD.2", 0)

		events, valid, err := e.VerifyAudit(id)
		if err != nil || !valid {
			t.Fatalf("audit must verify: %v", err)
		}
		if len(events) == 0 || events[0].Event != sharia.EventCreated {
			t.Fatal("expected created first")
		}
	})
}

func TestIjarahPenaltyToIncomeRejected(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "ijr_pen"
		fund(t, l, "@bank:treasury", "DZD.2", 10000000)
		mustCreateReq(t, e, ijarahReq(id))
		mustTransition(t, e, id, "acquire")
		mustTransition(t, e, id, "lease")
		_, err := e.Transition(id, "late_penalty", sharia.TransitionInput{
			Seq: 1, Amount: 20000, Destination: "@bank:income:ijarah",
		})
		se := shariaErr(t, err, sharia.ErrShariaViolation)
		if se.StandardRef != sharia.RefSS3 {
			t.Fatalf("expected SS-3, got %s", se.StandardRef)
		}
	})
}

func TestIjarahIdempotentRent(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "ijr_idem"
		fund(t, l, "@bank:treasury", "DZD.2", 10000000)
		fund(t, l, "@client:anis", "DZD.2", 2000000)
		mustCreateReq(t, e, ijarahReq(id))
		mustTransition(t, e, id, "acquire")
		mustTransition(t, e, id, "lease")
		if _, err := e.Transition(id, "pay_rent", sharia.TransitionInput{Seq: 1}); err != nil {
			t.Fatal(err)
		}
		_, err := e.Transition(id, "pay_rent", sharia.TransitionInput{Seq: 1})
		shariaErr(t, err, sharia.ErrDuplicate)
	})
}

package sharia_test

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

// rawParams marshals murabaha params into the json.RawMessage that
// CreateRequest now carries (the engine decodes per contract kind).
func rawParams(p sharia.MurabahaParams) json.RawMessage {
	raw, _ := json.Marshal(p)
	return raw
}

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	config.Init()
	viper.Set("storage.dir", os.TempDir())
	viper.Set("storage.sqlite.db_name", "sharia_engine")
	os.Remove(path.Join(os.TempDir(), "sharia_engine_engtest.db"))
	os.Exit(m.Run())
}

func withEngine(t *testing.T, f func(e *sharia.Engine, l *ledger.Ledger)) {
	t.Helper()
	fx.New(
		fx.NopLogger,
		fx.Provide(func(lc fx.Lifecycle) (*ledger.Ledger, error) {
			return ledger.NewLedger("engtest", lc)
		}),
		fx.Invoke(func(l *ledger.Ledger) {
			defer l.Close()
			f(sharia.NewEngine(l, l.Store()), l)
		}),
	)
}

func fund(t *testing.T, l *ledger.Ledger, account, asset string, amount int64) {
	t.Helper()
	_, err := l.Commit([]core.Transaction{{
		Postings: []core.Posting{
			{Source: core.WORLD, Destination: account, Asset: asset, Amount: amount},
		},
	}})
	if err != nil {
		t.Fatalf("failed to fund %s: %v", account, err)
	}
}

func balance(t *testing.T, l *ledger.Ledger, account, asset string) int64 {
	t.Helper()
	a, err := l.GetAccount(account)
	if err != nil {
		t.Fatal(err)
	}
	return a.Balances[asset]
}

func assertBal(t *testing.T, l *ledger.Ledger, account, asset string, expected int64) {
	t.Helper()
	if b := balance(t, l, account, asset); b != expected {
		t.Fatalf("wrong %s balance for %s: expected %d, got %d", asset, account, expected, b)
	}
}

func params(client, supplier, treasury string, installments int) sharia.MurabahaParams {
	return sharia.MurabahaParams{
		AssetCode:    "VHCL42A",
		Cost:         sharia.Monetary{Asset: "SAR2", Amount: 10000000},
		Markup:       sharia.Monetary{Asset: "SAR2", Amount: 1000000},
		Client:       client,
		Supplier:     supplier,
		BankTreasury: treasury,
		Installments: installments,
		FirstDue:     "2026-07-01T00:00:00Z",
		PeriodDays:   30,
	}
}

func shariaErr(t *testing.T, err error, code string) *sharia.Error {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %s, got nil", code)
	}
	se, ok := err.(*sharia.Error)
	if !ok {
		t.Fatalf("expected *sharia.Error, got %T: %v", err, err)
	}
	if se.Code != code {
		t.Fatalf("expected %s, got %s (%s)", code, se.Code, se.Message)
	}
	return se
}

func hasAuditEvent(events []sharia.AuditEvent, decision, transition string) bool {
	for _, e := range events {
		if e.Decision == decision && e.Transition == transition {
			return true
		}
	}
	return false
}

// Scenario A — nominal full cycle (spec §13).
func TestScenarioANominalCycle(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "mur_scen_a"
		treasury0 := balance(t, l, "@bank:treasury", "SAR2")
		income0 := balance(t, l, "@bank:income:murabaha", "SAR2")

		fund(t, l, "@bank:treasury", "SAR2", 20000000)

		c, schedule, err := e.Create(sharia.CreateRequest{
			Type:   sharia.TypeMurabaha,
			ID:     id,
			Params: rawParams(params("@client:ameziane", "@supplier:toyota", "@bank:treasury", 24)),
		})
		if err != nil {
			t.Fatal(err)
		}
		if c.State != sharia.StatePromise {
			t.Fatalf("expected PROMISE, got %s", c.State)
		}
		if len(schedule) != 24 || schedule[0].Amount != 458333 || schedule[23].Amount != 458341 {
			t.Fatalf("bad schedule: %d items", len(schedule))
		}

		// acquire
		res, err := e.Transition(id, "acquire", sharia.TransitionInput{})
		if err != nil {
			t.Fatal(err)
		}
		if res.NewState != sharia.StateAcquired {
			t.Fatalf("expected ACQUIRED, got %s", res.NewState)
		}
		assertBal(t, l, "@supplier:toyota", "SAR2", 10000000)
		assertBal(t, l, "@contracts:"+id+":inventory", "VHCL42A", 1)

		// sell
		res, err = e.Transition(id, "sell", sharia.TransitionInput{})
		if err != nil {
			t.Fatal(err)
		}
		if res.NewState != sharia.StateSold {
			t.Fatalf("expected SOLD, got %s", res.NewState)
		}
		assertBal(t, l, "@client:ameziane:assets", "VHCL42A", 1)
		assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", 11000000)
		assertBal(t, l, "@contracts:"+id+":deferred", "SAR2", 1000000)
		assertBal(t, l, "@contracts:"+id+":counterpart", "SAR2", -12000000)

		// pay all 24 installments, checking invariant §5.4 after each
		fund(t, l, "@client:ameziane", "SAR2", 11000000)
		var unpaidAmount, unpaidProfit int64 = 11000000, 1000000
		for seq := 1; seq <= 24; seq++ {
			res, err = e.Transition(id, "pay_installment", sharia.TransitionInput{Seq: seq})
			if err != nil {
				t.Fatalf("pay seq %d: %v", seq, err)
			}
			unpaidAmount -= schedule[seq-1].Amount
			unpaidProfit -= schedule[seq-1].ProfitPart
			assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", unpaidAmount)
			assertBal(t, l, "@contracts:"+id+":deferred", "SAR2", unpaidProfit)
		}

		if res.NewState != sharia.StateSettled {
			t.Fatalf("expected SETTLED after last payment, got %s", res.NewState)
		}
		assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", 0)
		assertBal(t, l, "@contracts:"+id+":deferred", "SAR2", 0)
		assertBal(t, l, "@bank:income:murabaha", "SAR2", income0+1000000)
		assertBal(t, l, "@bank:treasury", "SAR2", treasury0+20000000-10000000+11000000)

		// audit chain must verify
		events, valid, err := e.VerifyAudit(id)
		if err != nil {
			t.Fatal(err)
		}
		if !valid {
			t.Fatal("audit chain must be valid")
		}
		if len(events) == 0 || events[0].Event != sharia.EventCreated {
			t.Fatalf("expected created event first, got %+v", events)
		}
	})
}

// Scenario B — I-1 violation: sell without possession (qabd).
func TestScenarioBSellWithoutPossession(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "mur_scen_b"
		_, _, err := e.Create(sharia.CreateRequest{
			Type:   sharia.TypeMurabaha,
			ID:     id,
			Params: rawParams(params("@client:scen_b", "@supplier:scen_b", "@bank:treasury", 12)),
		})
		if err != nil {
			t.Fatal(err)
		}

		// Spec §5.3: the qabd check (I-1) runs FIRST — selling without
		// acquiring is a Sharia violation (422), not a mere FSM error.
		_, err = e.Transition(id, "sell", sharia.TransitionInput{})
		se := shariaErr(t, err, sharia.ErrShariaViolation)
		if se.StandardRef != sharia.RefSS8 {
			t.Fatalf("expected %s, got %s", sharia.RefSS8, se.StandardRef)
		}

		// no ledger transaction was created for this contract
		assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", 0)
		assertBal(t, l, "@contracts:"+id+":counterpart", "SAR2", 0)
		assertBal(t, l, "@client:scen_b:assets", "VHCL42A", 0)

		// the refusal is journalized (a refusal is evidence, not a silent error)
		events, valid, err := e.VerifyAudit(id)
		if err != nil || !valid {
			t.Fatalf("audit must verify: %v", err)
		}
		if !hasAuditEvent(events, sharia.DecisionDenied, "sell") {
			t.Fatal("expected a denied sell event in the audit trail")
		}
	})
}

// Scenario C — I-3 violation: penalty must never credit bank income.
func TestScenarioCPenaltyToIncomeRejected(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "mur_scen_c"
		income0 := balance(t, l, "@bank:income:murabaha", "SAR2")

		fund(t, l, "@bank:treasury", "SAR2", 10000000)
		mustCreate(t, e, id, params("@client:scen_c", "@supplier:scen_c", "@bank:treasury", 12))
		mustTransition(t, e, id, "acquire")
		mustTransition(t, e, id, "sell")

		_, err := e.Transition(id, "late_penalty", sharia.TransitionInput{
			Seq: 1, Amount: 20000, Destination: "@bank:income:murabaha",
		})
		se := shariaErr(t, err, sharia.ErrShariaViolation)
		if se.StandardRef != sharia.RefSS3 {
			t.Fatalf("expected %s, got %s", sharia.RefSS3, se.StandardRef)
		}
		assertBal(t, l, "@bank:income:murabaha", "SAR2", income0)

		// a lawful penalty to charity passes
		fund(t, l, "@client:scen_c", "SAR2", 20000)
		_, err = e.Transition(id, "late_penalty", sharia.TransitionInput{
			Seq: 1, Amount: 20000, Destination: "@charity:pool",
		})
		if err != nil {
			t.Fatal(err)
		}
		assertBal(t, l, "@charity:pool", "SAR2", 20000)
	})
}

// Scenario D — early settle with full rebate after 15 payments.
func TestScenarioDEarlySettle(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "mur_scen_d"
		income0 := balance(t, l, "@bank:income:murabaha", "SAR2")

		fund(t, l, "@bank:treasury", "SAR2", 10000000)
		_, schedule := mustCreate(t, e, id, params("@client:scen_d", "@supplier:scen_d", "@bank:treasury", 24))
		mustTransition(t, e, id, "acquire")
		mustTransition(t, e, id, "sell")

		var paidProfit int64
		fund(t, l, "@client:scen_d", "SAR2", 11000000)
		for seq := 1; seq <= 15; seq++ {
			if _, err := e.Transition(id, "pay_installment", sharia.TransitionInput{Seq: seq}); err != nil {
				t.Fatal(err)
			}
			paidProfit += schedule[seq-1].ProfitPart
		}

		// default rebate = full remaining profit
		res, err := e.Transition(id, "early_settle", sharia.TransitionInput{})
		if err != nil {
			t.Fatal(err)
		}
		if res.NewState != sharia.StateSettled {
			t.Fatalf("expected SETTLED, got %s", res.NewState)
		}

		assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", 0)
		assertBal(t, l, "@contracts:"+id+":deferred", "SAR2", 0)
		// income == profits of the 15 paid installments only (15 × 41666)
		assertBal(t, l, "@bank:income:murabaha", "SAR2", income0+paidProfit)
		if paidProfit != 15*41666 {
			t.Fatalf("expected paid profit 624990, got %d", paidProfit)
		}

		// remaining installments are settled_early
		items, err := l.Store().GetSchedule(id)
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range items {
			if it.Seq <= 15 && it.Status != sharia.StatusPaid {
				t.Fatalf("seq %d: expected paid, got %s", it.Seq, it.Status)
			}
			if it.Seq > 15 && it.Status != sharia.StatusSettledEarly {
				t.Fatalf("seq %d: expected settled_early, got %s", it.Seq, it.Status)
			}
		}
	})
}

// Scenario E — idempotence: replaying a payment is a clean failure.
func TestScenarioEIdempotence(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "mur_scen_e"

		fund(t, l, "@bank:treasury", "SAR2", 10000000)
		mustCreate(t, e, id, params("@client:scen_e", "@supplier:scen_e", "@bank:treasury", 12))
		mustTransition(t, e, id, "acquire")
		mustTransition(t, e, id, "sell")

		fund(t, l, "@client:scen_e", "SAR2", 2000000)
		if _, err := e.Transition(id, "pay_installment", sharia.TransitionInput{Seq: 1}); err != nil {
			t.Fatal(err)
		}

		clientBefore := balance(t, l, "@client:scen_e", "SAR2")
		receivableBefore := balance(t, l, "@contracts:"+id+":receivable", "SAR2")

		_, err := e.Transition(id, "pay_installment", sharia.TransitionInput{Seq: 1})
		shariaErr(t, err, sharia.ErrDuplicate)

		assertBal(t, l, "@client:scen_e", "SAR2", clientBefore)
		assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", receivableBefore)
	})
}

// Scenario F — FSM: out-of-sequence transitions are rejected AND journalized.
func TestScenarioFInvalidTransitions(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "mur_scen_f"

		fund(t, l, "@bank:treasury", "SAR2", 10000000)
		mustCreate(t, e, id, params("@client:scen_f", "@supplier:scen_f", "@bank:treasury", 12))

		// pay_installment from PROMISE
		_, err := e.Transition(id, "pay_installment", sharia.TransitionInput{Seq: 1})
		shariaErr(t, err, sharia.ErrInvalidTransition)

		mustTransition(t, e, id, "acquire")
		mustTransition(t, e, id, "sell")

		// cancel from SOLD (rule C-2)
		_, err = e.Transition(id, "cancel", sharia.TransitionInput{})
		shariaErr(t, err, sharia.ErrInvalidTransition)

		events, valid, err := e.VerifyAudit(id)
		if err != nil || !valid {
			t.Fatalf("audit must verify: %v", err)
		}
		if !hasAuditEvent(events, sharia.DecisionDenied, "pay_installment") {
			t.Fatal("expected denied pay_installment in audit")
		}
		if !hasAuditEvent(events, sharia.DecisionDenied, "cancel") {
			t.Fatal("expected denied cancel in audit")
		}
	})
}

// Unknown contract / unknown transition / acquire without treasury.
func TestEngineEdgeCases(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		_, err := e.Transition("mur_unknown", "acquire", sharia.TransitionInput{})
		shariaErr(t, err, sharia.ErrNotFound)

		const id = "mur_edge"
		// pauper treasury: acquire must fail on precondition
		mustCreate(t, e, id, params("@client:edge", "@supplier:edge", "@bank:pauper_treasury", 6))
		_, err = e.Transition(id, "acquire", sharia.TransitionInput{})
		shariaErr(t, err, sharia.ErrPrecondition)

		// invalid params at create
		_, _, err = e.Create(sharia.CreateRequest{
			Type:   sharia.TypeMurabaha,
			ID:     "mur_badparams",
			Params: rawParams(sharia.MurabahaParams{}),
		})
		shariaErr(t, err, sharia.ErrInvalidParams)

		// duplicate contract id
		_, _, err = e.Create(sharia.CreateRequest{
			Type:   sharia.TypeMurabaha,
			ID:     id,
			Params: rawParams(params("@client:edge", "@supplier:edge", "@bank:pauper_treasury", 6)),
		})
		shariaErr(t, err, sharia.ErrDuplicate)

		// bad type
		_, _, err = e.Create(sharia.CreateRequest{
			Type:   "ijarah",
			ID:     "mur_badtype",
			Params: rawParams(params("@client:edge", "@supplier:edge", "@bank:treasury", 6)),
		})
		shariaErr(t, err, sharia.ErrInvalidParams)
	})
}

// Cancel from PROMISE and from ACQUIRED.
func TestCancelFlows(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		// from PROMISE: no postings
		mustCreate(t, e, "mur_cxl_p", params("@client:cxl", "@supplier:cxl", "@bank:treasury", 6))
		res, err := e.Transition("mur_cxl_p", "cancel", sharia.TransitionInput{})
		if err != nil {
			t.Fatal(err)
		}
		if res.NewState != sharia.StateCancelled {
			t.Fatalf("expected CANCELLED, got %s", res.NewState)
		}
		if res.TxID != -1 {
			t.Fatalf("expected no ledger tx (-1), got %d", res.TxID)
		}

		// from ACQUIRED: asset moves to the bank's unsold inventory
		fund(t, l, "@bank:treasury", "SAR2", 10000000)
		unsold0 := balance(t, l, "@bank:inventory:unsold", "VHCL42A")
		mustCreate(t, e, "mur_cxl_a", params("@client:cxl", "@supplier:cxl", "@bank:treasury", 6))
		mustTransition(t, e, "mur_cxl_a", "acquire")
		res, err = e.Transition("mur_cxl_a", "cancel", sharia.TransitionInput{})
		if err != nil {
			t.Fatal(err)
		}
		if res.NewState != sharia.StateCancelled {
			t.Fatalf("expected CANCELLED, got %s", res.NewState)
		}
		assertBal(t, l, "@contracts:mur_cxl_a:inventory", "VHCL42A", 0)
		assertBal(t, l, "@bank:inventory:unsold", "VHCL42A", unsold0+1)
	})
}

func mustCreate(t *testing.T, e *sharia.Engine, id string, p sharia.MurabahaParams) (sharia.Contract, []sharia.Installment) {
	t.Helper()
	c, s, err := e.Create(sharia.CreateRequest{Type: sharia.TypeMurabaha, ID: id, Params: rawParams(p)})
	if err != nil {
		t.Fatal(err)
	}
	return c, s
}

func mustTransition(t *testing.T, e *sharia.Engine, id, name string) sharia.TransitionResult {
	t.Helper()
	res, err := e.Transition(id, name, sharia.TransitionInput{})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return res
}

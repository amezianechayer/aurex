package sharia_test

import (
	"encoding/json"
	"testing"

	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/sharia"
)

func TestGuardCoversContractEnginePostings(t *testing.T) {
	withEngine(t, func(e *sharia.Engine, l *ledger.Ledger) {
		const id = "mur_guard"
		fund(t, l, "@bank:treasury", "SAR2", 20000000)

		// a deny rule scoped to contract receivable accounts, capping any
		// posting above 1,000,000 — the sell will create an 11,000,000 receivable
		p, _ := json.Marshal(guard.AmountCapParams{
			Scope: "@contracts:*", Asset: "SAR2", Max: 1000000, Basis: "posting"})
		rule := guard.Rule{ID: "g-contracts", Kind: guard.KindAmountCap, Params: p,
			Action: guard.ActionDeny, Reason: "contract posting over cap",
			StandardRef: "POLICY-CONTRACT", Enabled: true,
			CreatedAt: "2026-06-13T00:00:00Z", UpdatedAt: "2026-06-13T00:00:00Z"}
		if err := l.Store().SaveRule(rule); err != nil {
			t.Fatal(err)
		}
		if err := l.Guard().Reload(); err != nil {
			t.Fatal(err)
		}

		mustCreate(t, e, id, params("@client:ameziane", "@supplier:toyota", "@bank:treasury", 24))
		mustTransition(t, e, id, "acquire")

		// sell builds an 11,000,000 receivable posting → guard must deny it
		_, err := e.Transition(id, "sell", sharia.TransitionInput{})
		if err == nil {
			t.Fatal("expected the guard rule to block the contract sell posting")
		}
		// the contract did not advance and no receivable was created
		assertBal(t, l, "@contracts:"+id+":receivable", "SAR2", 0)
	})
}

package ledger

import (
	"testing"

	"github.com/amezianechayer/corren/core"
)

func TestCounterpartAccountMayGoNegative(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()

		// The technical counterpart account of a sharia contract is the
		// ONLY non-world account allowed to carry a negative balance.
		_, err := l.Commit([]core.Transaction{{
			Postings: []core.Posting{
				{
					Source:      "@contracts:murvirt1:counterpart",
					Destination: "@contracts:murvirt1:receivable",
					Amount:      1000,
					Asset:       "SAR2",
				},
			},
		}})
		if err != nil {
			t.Fatalf("counterpart account must be exempt from balance check: %v", err)
		}

		assertBalance(t, l, "@contracts:murvirt1:counterpart", "SAR2", -1000)
		assertBalance(t, l, "@contracts:murvirt1:receivable", "SAR2", 1000)
	})
}

// Scenario G — an arbitrary FaRl script cannot drive a protected
// contract account negative: the standard balance check applies, only
// :counterpart is exempt.
func TestScriptCannotDrainProtectedAccounts(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()

		err := l.Execute(core.Script{
			Plain: `transfer [SAR2 5000] (
	from @contracts:murscript:receivable
	to   @attacker:wallet
)`,
		})
		if err == nil {
			t.Fatal("script must not be able to make receivable negative")
		}

		acc, err := l.GetAccount("@attacker:wallet")
		if err != nil {
			t.Fatal(err)
		}
		if acc.Balances["SAR2"] != 0 {
			t.Fatalf("attacker wallet must stay empty, got %d", acc.Balances["SAR2"])
		}
	})
}

func TestProtectedContractAccountsStayNonNegative(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()

		// receivable, deferred and inventory are NOT exempt (closed list).
		for _, src := range []string{
			"@contracts:murvirt2:receivable",
			"@contracts:murvirt2:deferred",
			"@contracts:murvirt2:inventory",
			"@contracts:murvirt2:counterpartx",
			"@bank:counterpart",
		} {
			_, err := l.Commit([]core.Transaction{{
				Postings: []core.Posting{
					{
						Source:      src,
						Destination: "@somewhere:else",
						Amount:      500,
						Asset:       "SAR2",
					},
				},
			}})
			if err == nil {
				t.Fatalf("account %s must NOT be exempt from the balance check", src)
			}
		}
	})
}

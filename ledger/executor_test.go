package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/amezianechayer/corren/core"
)

func assertBalance(t *testing.T, l *Ledger, account string, asset string, amount int64) {
	user, err := l.GetAccount(account)
	if err != nil {
		t.Error(err)
		return
	}
	if b := user.Balances[asset]; b != amount {
		t.Fatalf(
			"wrong %v balance for account %v, expected: %d got: %d",
			asset,
			account,
			amount,
			b,
		)
	}
}

func TestTransactionInvalidScript(t *testing.T) {
	with(func(l *Ledger) {
		script := core.Script{
			Plain: "this is not a valid script",
		}

		err := l.Execute(script)

		if err == nil {
			t.Error(errors.New(
				"script was invalid yet the transaction was commited",
			))
		}
		l.Close()
	})
}

func TestTransactionFail(t *testing.T) {
	with(func(l *Ledger) {
		script := core.Script{
			Plain: "fail",
		}

		err := l.Execute(script)

		if err == nil {
			t.Error(errors.New(
				"script failed yet the transaction was commited",
			))
		}
		l.Close()
	})
}

func TestSend(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()

		script := core.Script{
			Plain: `transfer [DZD.2 99] (
	from @world
	to   @user:001
)`,
		}

		err := l.Execute(script)

		if err != nil {
			t.Error(err)
			return
		}

		assertBalance(t, l, "@user:001", "DZD.2", 99)
	})
}

func TestVariables(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()

		var script core.Script
		json.Unmarshal(
			[]byte(`{
				"plain": "var $dest: account\ntransfer [DZD.2 42] (\n\tfrom @world\n\tto   $dest\n)",
				"vars": {
					"dest": "@user:042"
				}
			}`),
			&script,
		)

		err := l.Execute(script)

		if err != nil {
			t.Error(err)
			return
		}

		assertBalance(t, l, "@user:042", "DZD.2", 42)
	})
}

func TestEnoughFunds(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()

		tx := core.Transaction{
			Postings: []core.Posting{
				{
					Source:      "@world",
					Destination: "@user:001",
					Amount:      100,
					Asset:       "DZD.2",
				},
			},
		}

		_, err := l.Commit([]core.Transaction{tx})
		if err != nil {
			t.Error(err)
			return
		}

		var script core.Script
		json.Unmarshal(
			[]byte(`{
				"plain": "transfer [DZD.2 95] (\n\tfrom @user:001\n\tto   @world\n)"
			}`),
			&script,
		)

		err = l.Execute(script)
		if err != nil {
			t.Error(err)
			return
		}
	})
}

func TestNotEnoughFunds(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()

		tx := core.Transaction{
			Postings: []core.Posting{
				{
					Source:      "@world",
					Destination: "@user:002",
					Amount:      100,
					Asset:       "DZD.2",
				},
			},
		}

		_, err := l.Commit([]core.Transaction{tx})
		if err != nil {
			t.Error(err)
			return
		}

		var script core.Script
		json.Unmarshal(
			[]byte(`{
				"plain": "transfer [DZD.2 105] (\n\tfrom @user:002\n\tto   @world\n)"
			}`),
			&script,
		)

		err = l.Execute(script)
		if err == nil {
			t.Error("error wasn't supposed to be nil")
			return
		}
	})
}

func TestMetadata(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()

		tx := core.Transaction{
			Postings: []core.Posting{
				{
					Source:      "@world",
					Destination: "@sales:042",
					Amount:      100,
					Asset:       "DZD.2",
				},
			},
		}

		_, err := l.Commit([]core.Transaction{tx})
		if err != nil {
			t.Error(err)
			return
		}

		l.SaveMeta("account", "@sales:042", core.Metadata{
			"seller": json.RawMessage(`{
				"type":  "account",
				"value": "@users:053"
			}`),
		})

		l.SaveMeta("account", "@users:053", core.Metadata{
			"commission": json.RawMessage(`{
				"type":  "portion",
				"value": "15.5%"
			}`),
		})

		plain := `var $sale:       account
var $seller:     account = meta($sale, "seller")
var $commission: portion = meta($seller, "commission")

transfer [DZD.2 *] (
	from $sale
	to   {
		$commission to @plateforme
		remaining   to $seller
	}
)`

		script := core.Script{
			Plain: plain,
			Vars: map[string]json.RawMessage{
				"sale": json.RawMessage(`"@sales:042"`),
			},
		}

		err = l.Execute(script)
		if err != nil {
			t.Fatalf("execution error: %v", err)
		}

		// corren-vm >= 0.1.12 allocates 15.5% of 100 as 16 (rounds to
		// nearest); the remaining 84 goes to the seller
		assertBalance(t, l, "@sales:042", "DZD.2", 0)
		assertBalance(t, l, "@users:053", "DZD.2", 84)
		assertBalance(t, l, "@plateforme", "DZD.2", 16)

		fmt.Println("TestMetadata passed ✅")
	})
}

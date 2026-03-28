package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/amezianechayer/corren/core"
)

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
			Plain: "transfer [DZD.2 99] from @world to @user:001",
		}
		err := l.Execute(script)
		if err != nil {
			t.Error(err)
			return
		}
		user, err := l.GetAccount("@user:001")
		if err != nil {
			t.Error(err)
			return
		}
		if b := user.Balances["DZD.2"]; b != 99 {
			t.Error(fmt.Sprintf(
				"wrong DZD.2 balance for account @user:001, expected: %d got: %d",
				99,
				b,
			))
		}
	})
}

func TestVariables(t *testing.T) {
	with(func(l *Ledger) {
		defer l.Close()
		var script core.Script
		json.Unmarshal(
			[]byte(`{
				"plain": "var $dest: account\ntransfer [DZD.2 42] from @world to $dest\n",
				"vars": {
					"dest": "@user:042"
				}
			}`),
			&script)
		err := l.Execute(script)
		if err != nil {
			t.Error(err)
			return
		}
		user, err := l.GetAccount("@user:042")
		if err != nil {
			t.Error(err)
			return
		}
		if b := user.Balances["DZD.2"]; b != 42 {
			t.Error(fmt.Sprintf(
				"wrong DZD.2 balance for account @user:042, expected: %d got: %d",
				42,
				b,
			))
		}
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
		err := l.Commit([]core.Transaction{tx})
		if err != nil {
			t.Error(err)
			return
		}
		var script core.Script
		json.Unmarshal(
			[]byte(`{
				"plain": "transfer [DZD.2 95] from @user:001 to @world\n"
			}`),
			&script)
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
		err := l.Commit([]core.Transaction{tx})
		if err != nil {
			t.Error(err)
			return
		}
		var script core.Script
		json.Unmarshal(
			[]byte(`{
				"plain": "transfer [DZD.2 105] from @user:002 to @world\n"
			}`),
			&script)
		err = l.Execute(script)
		if err == nil {
			t.Error("error wasn't supposed to be nil")
			return
		}
	})
}

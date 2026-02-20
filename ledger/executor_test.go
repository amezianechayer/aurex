package ledger

import (
	"errors"
	"fmt"
	"testing"

	"github.com/amezianechayer/aurex/core"
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
	})
}

func TestSend(t *testing.T) {
	with(func(l *Ledger) {
		script := core.Script{
			Plain: "transfer [DZD.2 99] from @world to @user:001",
		}
		l.Execute(script)
		user, err := l.GetAccount("@user:001")
		if err != nil {
			t.Error(err)
		}
		if b := user.Balances["DZD.2"]; b != 99 {
			t.Error(fmt.Sprintf(
				"wrong DZD.2 balance for account @user:001, expected: %d got: %d",
				99,
				b,
			))
		}
		l.Close()
	})
}

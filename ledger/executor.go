package ledger

import (
	"errors"
	"fmt"

	"github.com/amezianechayer/corren-vm/script/compiler"
	"github.com/amezianechayer/corren-vm/vm"
	"github.com/amezianechayer/corren/core"
)

func (l *Ledger) Execute(script core.Script) error {
	if script.Plain == "" {
		return errors.New("no script to execute")
	}

	p, err := compiler.Compile(script.Plain)
	if err != nil {
		return fmt.Errorf("compile error: %v", err)
	}

	m := vm.NewMachine(p)

	err = m.SetVarsFromJSON(script.Vars)
	if err != nil {
		return fmt.Errorf("error while setting variables: %v", err)
	}

	{
		ch, err := m.ResolveResources()
		if err != nil {
			return fmt.Errorf("error while resolving resources: %v", err)
		}
		for req := range ch {
			if req.Error != nil {
				return fmt.Errorf("error in resource request: %v", req.Error)
			}
			req.Response <- nil
		}
	}

	{
		ch, err := m.ResolveBalances()
		if err != nil {
			return fmt.Errorf("error while resolving balances: %v", err)
		}
		for req := range ch {
			if req.Error != nil {
				return fmt.Errorf("error in balance request: %v", req.Error)
			}
			balances, err := l.store.AggregateBalances(req.Account)
			if err != nil {
				return fmt.Errorf("error fetching balance of %s: %v", req.Account, err)
			}
			amt := balances[req.Asset]
			if amt < 0 {
				amt = 0
			}
			req.Response <- uint64(amt)
		}
	}

	exit_code, err := m.Execute()
	if err != nil {
		return fmt.Errorf("script failed: %v", err)
	}
	if exit_code == vm.EXIT_FAIL {
		return errors.New("script exited with error code EXIT_FAIL")
	}

	t := core.Transaction{
		Postings: m.Postings,
	}

	return l.Commit([]core.Transaction{t})
}
package ledger

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/amezianechayer/corren-vm/script/compiler"
	"github.com/amezianechayer/corren-vm/vm"
	"github.com/amezianechayer/corren/core"
	"github.com/spf13/viper"
)

func validateGhararTerms(contract *core.ShariaContract) error {
	required, ok := core.RequiredTerms[contract.Type]
	if !ok {
		return nil
	}
	for _, field := range required {
		val, exists := contract.Terms[field]
		if !exists || val == nil || val == "" {
			return fmt.Errorf("gharar: contract %s (%s) is missing required term '%s'",
				contract.ID, contract.Type, field)
		}
	}
	if contract.Type == core.ContractQardHassan {
		for _, field := range []string{"markup", "interest", "profit", "return"} {
			if _, exists := contract.Terms[field]; exists {
				return fmt.Errorf("gharar: qard hassan contract %s must not contain profit term '%s' (riba prohibition)",
					contract.ID, field)
			}
		}
	}
	return nil
}

func (l *Ledger) Execute(script core.Script) error {
	if script.Plain == "" {
		return errors.New("no script to execute")
	}

	p, err := compiler.Compile(script.Plain)
	if err != nil {
		return fmt.Errorf("compile error: %v", err)
	}

	m := vm.NewMachine(p)

	if viper.GetBool("sharia.enabled") {
		if script.ContractID != "" {
			contract, err := l.store.FindContract(script.ContractID)
			if err != nil {
				return fmt.Errorf("could not load contract %s: %v", script.ContractID, err)
			}
			if contract != nil {
				if err := validateGhararTerms(contract); err != nil {
					return err
				}
			}
		}

		for _, name := range viper.GetStringSlice("sharia.constraints") {
			switch name {
			case "NO_RIBA":
				m.AddConstraint(&vm.NoRibaConstraint{})
			case "NO_GHARAR":
				m.AddConstraint(&vm.NoGhararConstraint{})
			case "NO_MAYSIR":
				m.AddConstraint(&vm.NoMaysirConstraint{})
			case "ASSET_BACKED":
				m.AddConstraint(vm.NewAssetBackedConstraint(viper.GetStringSlice("sharia.allowed_assets")))
			}
		}
	}

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
	switch exit_code {
	case vm.EXIT_FAIL:
		return errors.New("script exited with EXIT_FAIL")
	case vm.EXIT_FAIL_INVALID:
		return errors.New("script exited with invalid operation")
	case vm.EXIT_FAIL_INSUFFICIENT_FUNDS:
		return errors.New("insufficient funds")
	case vm.EXIT_FAIL_SHARIA_VIOLATION:
		if m.ShariaError != nil {
			return fmt.Errorf("sharia compliance failure: %v", m.ShariaError)
		}
		return errors.New("script exited with sharia violation")
	}

	ts := []core.Transaction{{Postings: m.Postings}}

	if err := l.Commit(ts); err != nil {
		return err
	}

	if script.ContractID != "" {
		now := time.Now().Format(time.RFC3339)
		raw := fmt.Sprintf("%s:%s:%d:%s", script.ContractID, script.Plain, ts[0].ID, now)
		h := sha256.Sum256([]byte(raw))
		cert := core.ShariaCertificate{
			ID:         fmt.Sprintf("%x", h),
			ContractID: script.ContractID,
			TxID:       ts[0].ID,
			Constraint: "SHARIA_COMPLIANT",
			Result:     "PASS",
			IssuedAt:   now,
			Authority:  "corren-ledger",
		}
		_ = l.SaveCertificate(cert)
	}

	return nil
}
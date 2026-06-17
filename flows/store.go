package flows

import (
	"github.com/amezianechayer/corren/core"
	sharia "github.com/amezianechayer/corren-sharia"
)

// FlowStore persists flow definitions and their instances. Implemented by
// storage/sqlite and storage/postgres, embedded into storage.Store.
type FlowStore interface {
	SaveFlow(f Flow) error
	UpdateFlow(f Flow) error
	GetFlow(id string) (Flow, error)
	ListFlows() ([]Flow, error)

	SaveInstance(i FlowInstance) error
	UpdateInstance(i FlowInstance) error
	GetInstance(id string) (FlowInstance, error)
	ListInstances(flowID string) ([]FlowInstance, error)
	ListInstancesByStatus(status string) ([]FlowInstance, error)
}

// LedgerPort is the ledger surface the transfer step and balance conditions use.
// *ledger.Ledger satisfies it.
type LedgerPort interface {
	Execute(core.Script) error
	GetAccount(address string) (core.Account, error)
}

// ShariaPort is the contract-engine surface the sharia step uses.
// *sharia.Engine satisfies it.
type ShariaPort interface {
	Create(req sharia.CreateRequest) (sharia.Contract, []sharia.Installment, error)
	Transition(contractID, name string, input sharia.TransitionInput) (sharia.TransitionResult, error)
}

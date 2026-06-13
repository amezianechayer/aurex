package sharia

import (
	"encoding/json"

	"github.com/amezianechayer/corren/core"
)

// Params is an opaque marker; each ContractKind knows its concrete type.
type Params interface{}

// Event is a transition request: a name plus the HTTP-provided input.
type Event struct {
	Name  string
	Input TransitionInput
}

// InstallmentMark is a declarative post-commit hook: mark one installment.
type InstallmentMark struct {
	Seq    int
	Status string
}

// TransitionPlan is what a kind asks the engine to commit, plus the
// declarative post-commit bookkeeping the engine performs generically.
type TransitionPlan struct {
	Postings    []core.Posting
	Reference   string
	NewState    string
	StandardRef string
	Event       string // audit event kind; defaults to EventTransition
	Payload     string
	Marks       []InstallmentMark // installments to mark paid/settled/etc.
	ExtraAudit  []AuditEvent      // additional audit events (e.g. settled)
}

// ContractKind defines one contract type. Adding a contract = adding a kind
// and registering it; the engine never changes.
type ContractKind interface {
	Type() string
	DecodeParams(raw json.RawMessage) (Params, error)
	ValidateParams(p Params) error
	BuildSchedule(p Params) ([]Installment, error)
	// ShariaGate runs BEFORE the FSM check so a possession/ownership
	// violation takes priority over a sequencing error. nil = no gate.
	ShariaGate(led LedgerPort, c Contract, p Params, ev Event) error
	AllowedTransitions(from string) []string
	Preconditions(led LedgerPort, c Contract, p Params, sched []Installment, ev Event) error
	BuildPlan(led LedgerPort, c Contract, p Params, sched []Installment, ev Event) (TransitionPlan, error)
}

var registry = map[string]ContractKind{}

func register(k ContractKind) { registry[k.Type()] = k }

func kindFor(contractType string) (ContractKind, bool) {
	k, ok := registry[contractType]
	return k, ok
}

// Package flows is a Sharia-native orchestration engine on top of the corren
// ledger, FaRl and the sharia contract engine. A Flow is a named sequence of
// steps run when a trigger fires; every money movement a step makes goes through
// the ledger (FaRl) and the sharia engine, so it always passes the Guard.
package flows

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"
)

// Triggers.
const (
	TriggerManual         = "manual"
	TriggerWalletCredited = "wallet_credited"
	TriggerSchedule       = "schedule"
)

// Step types.
const (
	StepTransfer  = "transfer"
	StepSharia    = "sharia"
	StepCondition = "condition"
	StepWait      = "wait"
)

// Instance statuses.
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusWaiting   = "waiting"   // paused on a wait step until ResumeAt
	StatusCompleted = "completed"
	StatusStopped   = "stopped"   // a condition evaluated false
	StatusFailed    = "failed"
)

// Flow is a stored orchestration definition.
type Flow struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Trigger   string          `json:"trigger"`
	Schedule  string          `json:"schedule,omitempty"` // when trigger=schedule
	Steps     json.RawMessage `json:"steps"`              // []Step
	LastRunAt string          `json:"last_run_at,omitempty"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// Step is one unit of a flow. Config is type-specific JSON.
type Step struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config,omitempty"`
}

// FlowInstance is one run of a flow.
type FlowInstance struct {
	ID          string          `json:"id"`
	FlowID      string          `json:"flow_id"`
	Status      string          `json:"status"`
	CurrentStep int             `json:"current_step"`
	Input       json.RawMessage `json:"input,omitempty"`
	ResumeAt    string          `json:"resume_at,omitempty"`
	Error       string          `json:"error,omitempty"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// --- step configs ---

type transferConfig struct {
	Script string `json:"script"`
}

type shariaConfig struct {
	Action       string          `json:"action"`        // "create" | "transition"
	ContractType string          `json:"contract_type"` // murabaha | ijarah
	ID           string          `json:"id"`
	Params       json.RawMessage `json:"params"`
	Transition   string          `json:"transition"`
}

type conditionConfig struct {
	Expr string `json:"expr"` // e.g. "amount >= 100000"
}

type waitConfig struct {
	Days int `json:"days"`
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func addDaysUTC(days int) string {
	return time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour).Format(time.RFC3339)
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

package sharia

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Audit event kinds
const (
	EventCreated    = "created"
	EventTransition = "transition"
	EventRejected   = "rejected"
	EventPenalty    = "penalty"
	EventSettled    = "settled"
	EventOverdue    = "overdue"

	DecisionAllowed = "allowed"
	DecisionDenied  = "denied"
)

type AuditEvent struct {
	ContractID  string `json:"contract_id"`
	Seq         int    `json:"seq"`
	Event       string `json:"event"`
	Transition  string `json:"transition"`
	Decision    string `json:"decision"`
	Reason      string `json:"reason"`
	StandardRef string `json:"standard_ref"`
	TxID        int64  `json:"tx_id"`
	Payload     string `json:"payload"`
	PrevHash    string `json:"prev_hash"`
	Hash        string `json:"hash"`
	CreatedAt   string `json:"created_at"`
}

// GenesisHash anchors each contract's chain: prev_hash(seq=0) = sha256(contract_id).
func GenesisHash(contractID string) string {
	h := sha256.Sum256([]byte(contractID))
	return hex.EncodeToString(h[:])
}

// ComputeAuditHash hashes prev_hash || canonical payload, where the canonical
// payload is compact JSON with sorted keys of the event WITHOUT hash fields.
func ComputeAuditHash(prevHash string, e AuditEvent) string {
	// json.Marshal sorts map keys, giving the canonical form.
	canonical, _ := json.Marshal(map[string]interface{}{
		"contract_id":  e.ContractID,
		"seq":          e.Seq,
		"event":        e.Event,
		"transition":   e.Transition,
		"decision":     e.Decision,
		"reason":       e.Reason,
		"standard_ref": e.StandardRef,
		"tx_id":        e.TxID,
		"payload":      e.Payload,
		"created_at":   e.CreatedAt,
	})

	h := sha256.New()
	h.Write([]byte(prevHash))
	h.Write(canonical)
	return hex.EncodeToString(h.Sum(nil))
}

// AppendChainedAudit chains an event onto the contract's audit trail
// (reads the last hash, links, persists) and returns the assigned seq.
func AppendChainedAudit(store ShariaStore, ev AuditEvent) (int, error) {
	seq, lastHash, err := store.LastAuditHash(ev.ContractID)
	if err != nil {
		return -1, err
	}
	prev := lastHash
	if seq == -1 {
		prev = GenesisHash(ev.ContractID)
	}
	ev.Seq = seq + 1
	ev.PrevHash = prev
	ev.Hash = ComputeAuditHash(prev, ev)
	return ev.Seq, store.AppendAudit(ev)
}

// VerifyChain re-walks a contract's full audit chain.
func VerifyChain(contractID string, events []AuditEvent) bool {
	prevHash := GenesisHash(contractID)
	for _, e := range events {
		if e.PrevHash != prevHash {
			return false
		}
		if ComputeAuditHash(prevHash, e) != e.Hash {
			return false
		}
		prevHash = e.Hash
	}
	return true
}

package sharia

import (
	"testing"
)

func makeChain(t *testing.T, contractID string, n int) []AuditEvent {
	t.Helper()
	events := []AuditEvent{}
	prevHash := GenesisHash(contractID)
	for i := 0; i < n; i++ {
		e := AuditEvent{
			ContractID:  contractID,
			Seq:         i,
			Event:       "transition",
			Transition:  "acquire",
			Decision:    "allowed",
			StandardRef: RefSS8,
			TxID:        int64(i),
			Payload:     `{"note":"event"}`,
			CreatedAt:   "2026-06-10T00:00:00Z",
		}
		e.PrevHash = prevHash
		e.Hash = ComputeAuditHash(prevHash, e)
		prevHash = e.Hash
		events = append(events, e)
	}
	return events
}

func TestAuditChainValid(t *testing.T) {
	events := makeChain(t, "mur20260610a7x4", 5)
	if !VerifyChain("mur20260610a7x4", events) {
		t.Fatal("expected chain to be valid")
	}
}

func TestAuditChainEmptyValid(t *testing.T) {
	if !VerifyChain("mur20260610a7x4", nil) {
		t.Fatal("empty chain must be valid")
	}
}

func TestAuditChainDetectsTamperedPayload(t *testing.T) {
	events := makeChain(t, "mur20260610a7x4", 5)
	events[2].Payload = `{"note":"falsified"}`
	if VerifyChain("mur20260610a7x4", events) {
		t.Fatal("expected tampered chain to be invalid")
	}
}

func TestAuditChainDetectsTamperedDecision(t *testing.T) {
	events := makeChain(t, "mur20260610a7x4", 3)
	events[1].Decision = "denied"
	if VerifyChain("mur20260610a7x4", events) {
		t.Fatal("expected tampered chain to be invalid")
	}
}

func TestAuditChainDetectsWrongGenesis(t *testing.T) {
	events := makeChain(t, "mur20260610a7x4", 3)
	if VerifyChain("another_contract", events) {
		t.Fatal("expected chain bound to another contract id to be invalid")
	}
}

func TestAuditHashIgnoresHashFields(t *testing.T) {
	e := AuditEvent{ContractID: "c", Seq: 0, Event: "created"}
	h1 := ComputeAuditHash("prev", e)
	e.PrevHash = "x"
	e.Hash = "y"
	h2 := ComputeAuditHash("prev", e)
	if h1 != h2 {
		t.Fatal("hash must not depend on PrevHash/Hash fields")
	}
}

func TestAuditHashDeterministic(t *testing.T) {
	e := AuditEvent{ContractID: "c", Seq: 1, Event: "transition", Payload: `{"a":1}`}
	if ComputeAuditHash("p", e) != ComputeAuditHash("p", e) {
		t.Fatal("hash must be deterministic")
	}
}

package guard

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/amezianechayer/corren/core"
)

type Engine struct {
	store GuardStore
	mu    sync.RWMutex
	rules []Rule
}

func NewEngine(store GuardStore) *Engine { return &Engine{store: store} }

func now() string { return time.Now().UTC().Format(time.RFC3339) }

// Reload refreshes the in-memory rule snapshot from the store. Called at
// startup and after every admin mutation (hot-reload).
func (e *Engine) Reload() error {
	rules, err := e.store.ListRules()
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.rules = rules
	e.mu.Unlock()
	return nil
}

func firstReference(txs []core.Transaction) string {
	for _, t := range txs {
		if t.Reference != "" {
			return t.Reference
		}
	}
	return ""
}

// Evaluate is the hook ledger.Commit calls just before SaveTransactions.
// On a matching deny rule it writes the deny guard_event immediately (its own
// store Exec — survives the rollback since SaveTransactions never runs) and
// returns *Error. Matching monitor rules are accumulated and returned for the
// ledger to persist AFTER a successful commit. No I/O unless a rule fires.
func (e *Engine) Evaluate(view LedgerView, txs []core.Transaction,
	netFlows map[string]map[string]int64) ([]GuardEvent, error) {

	e.mu.RLock()
	snapshot := e.rules
	e.mu.RUnlock()

	if len(snapshot) == 0 {
		return nil, nil
	}

	ref := firstReference(txs)
	var monitorEvents []GuardEvent

	for _, r := range snapshot {
		if !r.Enabled {
			continue
		}
		matched, detail := evalRule(r, txs, netFlows)
		if !matched {
			continue
		}
		payload, _ := json.Marshal(map[string]string{"detail": detail})
		ev := GuardEvent{
			RuleID: r.ID, Action: r.Action, Reason: r.Reason,
			StandardRef: r.StandardRef, TxReference: ref, Payload: string(payload),
			CreatedAt: now(),
		}
		if r.Action == ActionDeny {
			// proof must survive the (non-)commit: write now, then abort.
			if werr := e.store.AppendGuardEvent(ev); werr != nil {
				log.Printf("guard: failed to persist deny event for rule %s: %v", r.ID, werr)
			}
			return nil, &Error{Code: ErrGuardDenied, Message: r.Reason,
				StandardRef: r.StandardRef, RuleID: r.ID}
		}
		monitorEvents = append(monitorEvents, ev)
	}
	return monitorEvents, nil
}

// WriteMonitorEvents persists monitor events after a successful commit
// (best-effort, same durability model as the sharia audit trail).
func (e *Engine) WriteMonitorEvents(events []GuardEvent) {
	for _, ev := range events {
		if err := e.store.AppendGuardEvent(ev); err != nil {
			log.Printf("guard: failed to persist monitor event for rule %s: %v", ev.RuleID, err)
		}
	}
}

package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/sharia"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewScheduler),
	fx.Invoke(func(s *Scheduler) {}),
)

// Scheduler periodically marks overdue installments and journalizes them.
// It never posts to the ledger: detection is read + mark + audit only
// (spec §12). Penalties remain an explicit API decision in v1.
type Scheduler struct {
	resolver *ledger.Resolver
	stop     chan struct{}
}

func NewScheduler(lc fx.Lifecycle, resolver *ledger.Resolver) *Scheduler {
	viper.SetDefault("sharia.scheduler.interval", "1h")
	viper.SetDefault("sharia.grace_days", 7)

	s := &Scheduler{
		resolver: resolver,
		stop:     make(chan struct{}),
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go s.run()
			return nil
		},
		OnStop: func(context.Context) error {
			close(s.stop)
			return nil
		},
	})

	return s
}

func (s *Scheduler) run() {
	interval := viper.GetDuration("sharia.scheduler.interval")
	if interval <= 0 {
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *Scheduler) tick() {
	now := time.Now().UTC().Format(time.RFC3339)
	graceDays := viper.GetInt("sharia.grace_days")

	ledgers, _ := viper.Get("ledgers").([]interface{})
	for _, v := range ledgers {
		name, ok := v.(string)
		if !ok {
			continue
		}
		l, err := s.resolver.GetLedger(name)
		if err != nil {
			log.Printf("sharia scheduler: ledger %s: %v\n", name, err)
			continue
		}
		marked, err := RunOnce(l, now, graceDays)
		if err != nil {
			log.Printf("sharia scheduler: ledger %s: %v\n", name, err)
			continue
		}
		if marked > 0 {
			log.Printf("sharia scheduler: ledger %s: %d installment(s) marked overdue\n", name, marked)
		}
	}
}

// RunOnce scans one ledger for pending installments past their grace
// period on SOLD contracts, marks them overdue and journalizes each.
// Returns the number of installments marked.
func RunOnce(l *ledger.Ledger, nowRFC3339 string, graceDays int) (int, error) {
	store := l.Store()

	due, err := store.FindDue(nowRFC3339, graceDays)
	if err != nil {
		return 0, err
	}

	marked := 0
	for _, item := range due {
		c, err := store.GetContract(item.ContractID)
		if err != nil {
			return marked, err
		}
		if c.State != sharia.StateSold {
			continue
		}

		if err := store.MarkInstallment(item.ContractID, item.Seq, sharia.StatusOverdue, -1, ""); err != nil {
			return marked, err
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"seq": item.Seq, "due_date": item.DueDate, "amount": item.Amount, "grace_days": graceDays,
		})
		if _, err := sharia.AppendChainedAudit(store, sharia.AuditEvent{
			ContractID:  item.ContractID,
			Event:       sharia.EventOverdue,
			Decision:    sharia.DecisionAllowed,
			StandardRef: sharia.RefSS3,
			TxID:        -1,
			Payload:     string(payload),
			CreatedAt:   nowRFC3339,
		}); err != nil {
			return marked, err
		}
		marked++
	}

	return marked, nil
}

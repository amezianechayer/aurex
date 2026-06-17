package scheduler

import (
	"context"
	"log"
	"time"

	csharia "github.com/amezianechayer/corren-sharia"
	"github.com/amezianechayer/corren/flows"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/shariawire"
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

		// Orchestration: fire due scheduled flows and resume waiting instances.
		if she, eerr := shariawire.EngineFor(l); eerr == nil {
			flows.NewEngine(l, she, l.Store()).RunDue(now)
		}
	}
}

// RunOnce marks overdue installments for one ledger by delegating to the
// corren-sharia engine on its store (same database). Thin wrapper: the sharia
// scheduling logic lives in corren-sharia.RunOverdue.
func RunOnce(l *ledger.Ledger, nowRFC3339 string, graceDays int) (int, error) {
	store, err := shariawire.StoreFor(l)
	if err != nil {
		return 0, err
	}
	return csharia.RunOverdue(store, nowRFC3339, graceDays)
}

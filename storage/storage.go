package storage

import (
	"github.com/amezianechayer/corren/analytics"
	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/flows"
	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger/query"
	"github.com/amezianechayer/corren/payments"
	"github.com/amezianechayer/corren/sharia"
	"github.com/amezianechayer/corren/storage/postgres"
	"github.com/amezianechayer/corren/storage/sqlite"
	"github.com/amezianechayer/corren/wallets"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Store interface {
	SaveTransactions([]core.Transaction) error
	CountTransactions() (int64, error)
	FindTransactions(query.Query) (query.Cursor, error)
	GetTransaction(string) (core.Transaction, error)
	AggregateBalances(string) (map[string]int64, error)
	AggregateVolumes(string) (map[string]map[string]int64, error)
	CountAccounts() (int64, error)
	FindAccounts(query.Query) (query.Cursor, error)
	SaveMeta(string, string, string, string, string, string) error

	CountMeta() (int64, error)
	GetMeta(string, string) (core.Metadata, error)
	Initialize() error
	Close()

	sharia.ShariaStore
	guard.GuardStore
	analytics.AnalyticsStore
	wallets.WalletStore
	payments.PaymentStore
	flows.FlowStore
}

func GetStore(name string) (Store, error) {
	driver := viper.GetString("storage.driver")

	switch driver {
	case "sqlite":
		return sqlite.NewStore(name)
	case "postgres":
		return postgres.NewStore(name)
	default:
		break
	}

	panic(errors.Errorf(
		"unsupported store: %s",
		driver,
	))
}

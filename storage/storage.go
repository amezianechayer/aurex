package storage

import (
	"github.com/amezianechayer/aurex/config"
	"github.com/amezianechayer/aurex/core"
	"github.com/amezianechayer/aurex/ledger/query"
	"github.com/amezianechayer/aurex/storage/postgres"
	"github.com/amezianechayer/aurex/storage/sqlite"
	"github.com/pkg/errors"
)

type Store interface {
	SaveTransactions([]core.Transaction) error
	CountTransactions() (int64, error)
	FindTransactions(query.Query) (query.Cursor, error)
	AggregateBalances(string) (map[string]int64, error)
	CountAccounts() (int64, error)
	FindAccounts(query.Query) (query.Cursor, error)
	Initialize() error
	Close()
}

func GetStore(c config.Config) (Store, error) {
	switch c.Storage.Driver {
	case "sqlite":
		return sqlite.NewStore(c)
	case "postgres":
		return postgres.NewStore(c)
	default:
		break
	}

	panic(errors.Errorf(
		"unsupported store: %s",
		c.Storage.Driver,
	))
}

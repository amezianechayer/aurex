package storage

import (
	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger/query"
	"github.com/amezianechayer/corren/storage/postgres"
	"github.com/amezianechayer/corren/storage/sqlite"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Store interface {
	SaveTransactions([]core.Transaction) error
	CountTransactions() (int64, error)
	FindTransactions(query.Query) (query.Cursor, error)
	AggregateBalances(string) (map[string]int64, error)
	AggregateVolumes(string) (map[string]map[string]int64, error)
	CountAccounts() (int64, error)
	FindAccounts(query.Query) (query.Cursor, error)
	Initialize() error
	Close()

	// Asset registry
	SaveAsset(core.AssetEntry) error
	FindAssets() ([]core.AssetEntry, error)
	FindAsset(string) (*core.AssetEntry, error)

	// Sharia contracts
	SaveContract(core.ShariaContract) error
	UpdateContractStatus(id string, status core.ContractStatus) error
	FindContracts(query.Query) (query.Cursor, error)
	FindContract(string) (*core.ShariaContract, error)

	// Sharia certificates
	SaveCertificate(core.ShariaCertificate) error
	FindCertificates(contractID string) ([]core.ShariaCertificate, error)

	// API keys
	SaveAPIKey(core.APIKey) error
	FindAPIKey(keyHash string) (*core.APIKey, error)
	DeleteAPIKey(keyHash string) error
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

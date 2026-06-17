// Package shariawire is the host-side glue between corren's ledger and the
// standalone corren-sharia module: an adapter that lets the ledger satisfy
// corren-sharia's chain-typed LedgerPort, and a per-ledger engine factory backed
// by a corren-sharia store on the same database. All sharia LOGIC lives in
// corren-sharia; this package is wiring only.
package shariawire

import (
	"fmt"
	"path"
	"sync"

	csharia "github.com/amezianechayer/corren-sharia"
	"github.com/amezianechayer/corren-sharia/chain"
	cspg "github.com/amezianechayer/corren-sharia/storage/postgres"
	cssqlite "github.com/amezianechayer/corren-sharia/storage/sqlite"
	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger"
	"github.com/spf13/viper"
)

// shariaLedger adapts *ledger.Ledger to corren-sharia's chain-typed LedgerPort,
// converting between corren/core and corren-sharia/chain (structurally identical
// value types).
type shariaLedger struct{ l *ledger.Ledger }

func (a shariaLedger) Commit(ts []chain.Transaction) ([]chain.Transaction, error) {
	in := make([]core.Transaction, len(ts))
	for i, t := range ts {
		in[i] = toCore(t)
	}
	out, err := a.l.Commit(in)
	if err != nil {
		return nil, err
	}
	res := make([]chain.Transaction, len(out))
	for i, t := range out {
		res[i] = toChain(t)
	}
	return res, nil
}

func (a shariaLedger) GetAccount(addr string) (chain.Account, error) {
	acc, err := a.l.GetAccount(addr)
	if err != nil {
		return chain.Account{}, err
	}
	return chain.Account{Address: acc.Address, Contract: acc.Contract, Type: acc.Type, Balances: acc.Balances, Metadata: chain.Metadata(acc.Metadata)}, nil
}

func (a shariaLedger) SaveMeta(targetType, targetID string, m chain.Metadata) error {
	return a.l.SaveMeta(targetType, targetID, core.Metadata(m))
}

func toCore(t chain.Transaction) core.Transaction {
	ps := make(core.Postings, len(t.Postings))
	for i, p := range t.Postings {
		ps[i] = core.Posting{Source: p.Source, Destination: p.Destination, Amount: p.Amount, Asset: p.Asset}
	}
	return core.Transaction{ID: t.ID, Postings: ps, Reference: t.Reference, Timestamp: t.Timestamp, Hash: t.Hash, Metadata: core.Metadata(t.Metadata)}
}

func toChain(t core.Transaction) chain.Transaction {
	ps := make(chain.Postings, len(t.Postings))
	for i, p := range t.Postings {
		ps[i] = chain.Posting{Source: p.Source, Destination: p.Destination, Amount: p.Amount, Asset: p.Asset}
	}
	return chain.Transaction{ID: t.ID, Postings: ps, Reference: t.Reference, Timestamp: t.Timestamp, Hash: t.Hash, Metadata: chain.Metadata(t.Metadata)}
}

var (
	mu      sync.Mutex
	engines = map[string]*csharia.Engine{}
)

// EngineFor returns the corren-sharia engine for a ledger (cached per ledger so
// the contract locks and the store connection survive across requests), backed
// by a corren-sharia store on the same database plus the ledger adapter.
func EngineFor(l *ledger.Ledger) (*csharia.Engine, error) {
	mu.Lock()
	defer mu.Unlock()
	if e, ok := engines[l.Name()]; ok {
		return e, nil
	}
	store, err := newStore(l.Name())
	if err != nil {
		return nil, err
	}
	e := csharia.NewEngine(shariaLedger{l}, store)
	engines[l.Name()] = e
	return e, nil
}

// newStore builds a corren-sharia store on the same database as the ledger
// (own tables + migrations, same DB).
func newStore(ledgerName string) (csharia.ShariaStore, error) {
	if viper.GetString("storage.driver") == "postgres" {
		return cspg.NewStore(viper.GetString("storage.postgres.conn_string"), ledgerName)
	}
	dbpath := fmt.Sprintf("file:%s?_journal=WAL", path.Join(
		viper.GetString("storage.dir"),
		fmt.Sprintf("%s_%s.db", viper.GetString("storage.sqlite.db_name"), ledgerName),
	))
	return cssqlite.NewStore(dbpath)
}

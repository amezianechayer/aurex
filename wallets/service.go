package wallets

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/amezianechayer/corren/core"
)

// Service is the wallet engine. It owns no balances: every movement is a ledger
// transaction committed through the LedgerPort, so it always passes the Guard,
// and every balance is read back from the ledger.
type Service struct {
	led   LedgerPort
	store WalletStore
}

func NewService(led LedgerPort, store WalletStore) *Service {
	return &Service{led: led, store: store}
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

// Create registers a new wallet. No ledger posting happens — a fresh wallet is
// simply empty until credited.
func (s *Service) Create(owner, asset string) (Wallet, error) {
	if owner == "" {
		return Wallet{}, newError(ErrInvalidParams, "owner is required")
	}
	if asset == "" {
		return Wallet{}, newError(ErrInvalidParams, "asset is required")
	}
	ts := now()
	w := Wallet{ID: generateID("wlt_"), Owner: owner, Asset: asset, CreatedAt: ts, UpdatedAt: ts}
	if err := s.store.SaveWallet(w); err != nil {
		return Wallet{}, err
	}
	return w, nil
}

func (s *Service) Get(id string) (Wallet, error) {
	w, err := s.store.GetWallet(id)
	if err != nil {
		return Wallet{}, newError(ErrNotFound, "wallet not found")
	}
	return w, nil
}

func (s *Service) List(limit, offset int) ([]Wallet, error) {
	return s.store.ListWallets(limit, offset)
}

// Balances reads available (main) and held (holds) amounts straight from the
// ledger, per asset.
func (s *Service) Balances(id string) (Balances, error) {
	if _, err := s.Get(id); err != nil {
		return Balances{}, err
	}
	main, err := s.led.GetAccount(MainAccount(id))
	if err != nil {
		return Balances{}, err
	}
	holds, err := s.led.GetAccount(HoldsAccount(id))
	if err != nil {
		return Balances{}, err
	}
	return Balances{
		WalletID:  id,
		Available: nonNil(main.Balances),
		Held:      nonNil(holds.Balances),
	}, nil
}

// Credit funds the wallet (topup). source defaults to @world (external money
// entering the system). The posting passes the Guard at Commit.
func (s *Service) Credit(id, asset string, amount int64, source string) (core.Transaction, error) {
	w, err := s.requireWallet(id, &asset)
	if err != nil {
		return core.Transaction{}, err
	}
	if err := requirePositive(amount); err != nil {
		return core.Transaction{}, err
	}
	if source == "" {
		source = core.WORLD
	}
	return s.move(source, MainAccount(w.ID), asset, amount, "credit", w.ID)
}

// Debit moves funds out of the wallet. destination defaults to @world (cash-out
// sink); for a P2P transfer, pass the recipient's main account. Insufficient
// available funds are rejected by the ledger and surfaced as
// ERR_INSUFFICIENT_FUNDS.
func (s *Service) Debit(id, asset string, amount int64, destination string) (core.Transaction, error) {
	w, err := s.requireWallet(id, &asset)
	if err != nil {
		return core.Transaction{}, err
	}
	if err := requirePositive(amount); err != nil {
		return core.Transaction{}, err
	}
	if destination == "" {
		destination = core.WORLD
	}
	return s.move(MainAccount(w.ID), destination, asset, amount, "debit", w.ID)
}

// Transfer moves funds from one wallet to another: a single posting
// fromMain -> toMain. Both wallets must exist; insufficient funds are refused by
// the ledger (no implicit debt = no riba).
func (s *Service) Transfer(fromID, toID, asset string, amount int64) (core.Transaction, error) {
	from, err := s.requireWallet(fromID, &asset)
	if err != nil {
		return core.Transaction{}, err
	}
	to, err := s.store.GetWallet(toID)
	if err != nil {
		return core.Transaction{}, &Error{Code: ErrNotFound, Message: "destination wallet not found", WalletID: toID}
	}
	if err := requirePositive(amount); err != nil {
		return core.Transaction{}, err
	}
	return s.move(MainAccount(from.ID), MainAccount(to.ID), asset, amount, "transfer", from.ID)
}

// Hold reserves funds: a real posting main -> holds. Sharia-aware by
// construction — the ledger refuses it if the available balance is short, so a
// hold can never create fictitious debt.
func (s *Service) Hold(id, asset string, amount int64, reason, expiresAt string) (Hold, error) {
	w, err := s.requireWallet(id, &asset)
	if err != nil {
		return Hold{}, err
	}
	if err := requirePositive(amount); err != nil {
		return Hold{}, err
	}
	if _, err := s.move(MainAccount(w.ID), HoldsAccount(w.ID), asset, amount, "hold", w.ID); err != nil {
		return Hold{}, err
	}
	ts := now()
	h := Hold{
		ID: generateID("hld_"), WalletID: w.ID, Asset: asset, Amount: amount,
		Status: HoldActive, Reason: reason, ExpiresAt: expiresAt, CreatedAt: ts, UpdatedAt: ts,
	}
	if err := s.store.SaveHold(h); err != nil {
		return Hold{}, err
	}
	return h, nil
}

// Release frees a hold back to the wallet's available balance (holds -> main).
// This is the DELETE /holds/:id operation.
func (s *Service) Release(walletID, holdID string) (core.Transaction, error) {
	return s.VoidHold(walletID, holdID)
}

// CaptureHold settles a hold to a destination (default @world): holds ->
// destination, then marks the hold captured.
func (s *Service) CaptureHold(walletID, holdID, destination string) (core.Transaction, error) {
	h, err := s.requireActiveHold(walletID, holdID)
	if err != nil {
		return core.Transaction{}, err
	}
	if destination == "" {
		destination = core.WORLD
	}
	tx, err := s.move(HoldsAccount(walletID), destination, h.Asset, h.Amount, "capture_hold", walletID)
	if err != nil {
		return core.Transaction{}, err
	}
	h.Status = HoldCaptured
	h.UpdatedAt = now()
	if err := s.store.UpdateHold(h); err != nil {
		return core.Transaction{}, err
	}
	return tx, nil
}

// VoidHold releases a hold back to the wallet: holds -> main, then marks it
// voided.
func (s *Service) VoidHold(walletID, holdID string) (core.Transaction, error) {
	h, err := s.requireActiveHold(walletID, holdID)
	if err != nil {
		return core.Transaction{}, err
	}
	tx, err := s.move(HoldsAccount(walletID), MainAccount(walletID), h.Asset, h.Amount, "void_hold", walletID)
	if err != nil {
		return core.Transaction{}, err
	}
	h.Status = HoldVoided
	h.UpdatedAt = now()
	if err := s.store.UpdateHold(h); err != nil {
		return core.Transaction{}, err
	}
	return tx, nil
}

func (s *Service) ListHolds(walletID string) ([]Hold, error) {
	if _, err := s.Get(walletID); err != nil {
		return nil, err
	}
	return s.store.ListHolds(walletID)
}

// move commits a single-posting transaction and tags it for the audit trail.
// All wallet money movement funnels through here — and thus through the Guard.
func (s *Service) move(from, to, asset string, amount int64, op, walletID string) (core.Transaction, error) {
	tx := core.Transaction{
		Postings: []core.Posting{{Source: from, Destination: to, Asset: asset, Amount: amount}},
		Reference: fmt.Sprintf("wallet:%s:%s:%d", op, walletID, time.Now().UnixNano()),
	}
	committed, err := s.led.Commit([]core.Transaction{tx})
	if err != nil {
		return core.Transaction{}, mapCommitErr(err)
	}
	out := committed[0]
	// Best-effort audit tag; never fail the (already committed) movement on it.
	_ = s.led.SaveMeta("transaction", fmt.Sprint(out.ID), core.Metadata{
		"wallet/operation": json.RawMessage(fmt.Sprintf("%q", op)),
		"wallet/wallet_id": json.RawMessage(fmt.Sprintf("%q", walletID)),
	})
	return out, nil
}

func (s *Service) requireWallet(id string, asset *string) (Wallet, error) {
	w, err := s.Get(id)
	if err != nil {
		return Wallet{}, err
	}
	if *asset == "" {
		*asset = w.Asset
	}
	return w, nil
}

func (s *Service) requireActiveHold(walletID, holdID string) (Hold, error) {
	if _, err := s.Get(walletID); err != nil {
		return Hold{}, err
	}
	h, err := s.store.GetHold(holdID)
	if err != nil || h.WalletID != walletID {
		return Hold{}, newError(ErrNotFound, "hold not found")
	}
	if h.Status != HoldActive {
		return Hold{}, &Error{Code: ErrHoldNotActive, Message: "hold is " + h.Status, HoldID: holdID, WalletID: walletID}
	}
	return h, nil
}

func requirePositive(amount int64) error {
	if amount <= 0 {
		return newError(ErrInvalidParams, "amount must be > 0")
	}
	return nil
}

// mapCommitErr turns the ledger's insufficient-balance error into a typed
// wallet error. Other Commit errors (e.g. a *guard.Error deny) are returned
// as-is so the controller can render them with their own HTTP status.
func mapCommitErr(err error) error {
	if err == nil {
		return nil
	}
	if strings.HasPrefix(err.Error(), "balance.insufficient") {
		return newError(ErrInsufficientFunds, "insufficient wallet balance")
	}
	return err
}

func nonNil(m map[string]int64) map[string]int64 {
	if m == nil {
		return map[string]int64{}
	}
	return m
}

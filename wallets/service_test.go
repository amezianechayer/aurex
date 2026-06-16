package wallets

import (
	"fmt"
	"testing"

	"github.com/amezianechayer/corren/core"
)

// fakeLedger is an in-memory LedgerPort that enforces the same non-negative
// invariant as the real ledger for non-@world accounts, so insufficient-funds
// and Sharia-aware hold behaviour are exercised for real.
type fakeLedger struct {
	bal map[string]map[string]int64
	seq int64
}

func newFakeLedger() *fakeLedger { return &fakeLedger{bal: map[string]map[string]int64{}} }

func (f *fakeLedger) balance(acct, asset string) int64 {
	if f.bal[acct] == nil {
		return 0
	}
	return f.bal[acct][asset]
}

func (f *fakeLedger) add(acct, asset string, delta int64) {
	if f.bal[acct] == nil {
		f.bal[acct] = map[string]int64{}
	}
	f.bal[acct][asset] += delta
}

func (f *fakeLedger) Commit(ts []core.Transaction) ([]core.Transaction, error) {
	for _, t := range ts {
		for _, p := range t.Postings {
			if p.Source != core.WORLD && f.balance(p.Source, p.Asset)-p.Amount < 0 {
				return nil, fmt.Errorf("balance.insufficient.%s", p.Asset)
			}
		}
	}
	out := make([]core.Transaction, 0, len(ts))
	for _, t := range ts {
		for _, p := range t.Postings {
			if p.Source != core.WORLD {
				f.add(p.Source, p.Asset, -p.Amount)
			}
			f.add(p.Destination, p.Asset, p.Amount)
		}
		f.seq++
		t.ID = f.seq
		out = append(out, t)
	}
	return out, nil
}

func (f *fakeLedger) GetAccount(address string) (core.Account, error) {
	return core.Account{Address: address, Balances: f.bal[address]}, nil
}

func (f *fakeLedger) SaveMeta(string, string, core.Metadata) error { return nil }

// memStore is an in-memory WalletStore.
type memStore struct {
	wallets map[string]Wallet
	holds   map[string]Hold
}

func newMemStore() *memStore {
	return &memStore{wallets: map[string]Wallet{}, holds: map[string]Hold{}}
}

func (m *memStore) SaveWallet(w Wallet) error { m.wallets[w.ID] = w; return nil }
func (m *memStore) GetWallet(id string) (Wallet, error) {
	w, ok := m.wallets[id]
	if !ok {
		return Wallet{}, fmt.Errorf("not found")
	}
	return w, nil
}
func (m *memStore) ListWallets(limit, offset int) ([]Wallet, error) {
	out := []Wallet{}
	for _, w := range m.wallets {
		out = append(out, w)
	}
	return out, nil
}
func (m *memStore) SaveHold(h Hold) error   { m.holds[h.ID] = h; return nil }
func (m *memStore) UpdateHold(h Hold) error { m.holds[h.ID] = h; return nil }
func (m *memStore) GetHold(id string) (Hold, error) {
	h, ok := m.holds[id]
	if !ok {
		return Hold{}, fmt.Errorf("not found")
	}
	return h, nil
}
func (m *memStore) ListHolds(walletID string) ([]Hold, error) {
	out := []Hold{}
	for _, h := range m.holds {
		if h.WalletID == walletID {
			out = append(out, h)
		}
	}
	return out, nil
}

func newService() (*Service, *fakeLedger) {
	led := newFakeLedger()
	return NewService(led, newMemStore()), led
}

const asset = "AED.2"

func TestCreateValidation(t *testing.T) {
	svc, _ := newService()
	if _, err := svc.Create("", asset); err == nil {
		t.Fatal("expected error on empty owner")
	}
	if _, err := svc.Create("@user:alice", ""); err == nil {
		t.Fatal("expected error on empty asset")
	}
	w, err := svc.Create("@user:alice", asset)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if w.ID == "" || w.Owner != "@user:alice" || w.Asset != asset {
		t.Fatalf("unexpected wallet %+v", w)
	}
}

func TestCreditDebitFlow(t *testing.T) {
	svc, _ := newService()
	w, _ := svc.Create("@user:alice", asset)

	if _, err := svc.Credit(w.ID, "", 100000, ""); err != nil { // 1000.00 AED, default asset & source @world
		t.Fatalf("credit: %v", err)
	}
	b, _ := svc.Balances(w.ID)
	if b.Available[asset] != 100000 {
		t.Fatalf("expected available 100000, got %d", b.Available[asset])
	}

	if _, err := svc.Debit(w.ID, "", 20000, ""); err != nil { // 200.00 out
		t.Fatalf("debit: %v", err)
	}
	b, _ = svc.Balances(w.ID)
	if b.Available[asset] != 80000 {
		t.Fatalf("expected available 80000, got %d", b.Available[asset])
	}
}

func TestDebitInsufficient(t *testing.T) {
	svc, _ := newService()
	w, _ := svc.Create("@user:alice", asset)
	svc.Credit(w.ID, "", 5000, "")

	_, err := svc.Debit(w.ID, "", 9000, "")
	we, ok := err.(*Error)
	if !ok || we.Code != ErrInsufficientFunds {
		t.Fatalf("expected ERR_INSUFFICIENT_FUNDS, got %v", err)
	}
}

func TestAmountMustBePositive(t *testing.T) {
	svc, _ := newService()
	w, _ := svc.Create("@user:alice", asset)
	if _, err := svc.Credit(w.ID, "", 0, ""); err == nil {
		t.Fatal("expected error on non-positive amount")
	}
}

func TestHoldLifecycle(t *testing.T) {
	svc, _ := newService()
	w, _ := svc.Create("@user:alice", asset)
	svc.Credit(w.ID, "", 100000, "")

	h, err := svc.Hold(w.ID, "", 30000, "pending purchase", "2026-12-31T00:00:00Z")
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	b, _ := svc.Balances(w.ID)
	if b.Available[asset] != 70000 || b.Held[asset] != 30000 {
		t.Fatalf("expected available 70000 / held 30000, got %d / %d", b.Available[asset], b.Held[asset])
	}

	if _, err := svc.CaptureHold(w.ID, h.ID, ""); err != nil { // settle to @world
		t.Fatalf("capture: %v", err)
	}
	b, _ = svc.Balances(w.ID)
	if b.Held[asset] != 0 || b.Available[asset] != 70000 {
		t.Fatalf("after capture expected held 0 / available 70000, got %d / %d", b.Held[asset], b.Available[asset])
	}

	// a second hold, then void it → funds return to available
	h2, _ := svc.Hold(w.ID, "", 10000, "", "")
	if _, err := svc.VoidHold(w.ID, h2.ID); err != nil {
		t.Fatalf("void: %v", err)
	}
	b, _ = svc.Balances(w.ID)
	if b.Available[asset] != 70000 || b.Held[asset] != 0 {
		t.Fatalf("after void expected available 70000 / held 0, got %d / %d", b.Available[asset], b.Held[asset])
	}
}

func TestHoldShariaAwareNoFictitiousDebt(t *testing.T) {
	svc, _ := newService()
	w, _ := svc.Create("@user:alice", asset)
	svc.Credit(w.ID, "", 5000, "")

	// holding more than available must be refused by the ledger invariant
	_, err := svc.Hold(w.ID, "", 9000, "", "")
	we, ok := err.(*Error)
	if !ok || we.Code != ErrInsufficientFunds {
		t.Fatalf("expected ERR_INSUFFICIENT_FUNDS (no hold on funds that don't exist), got %v", err)
	}
}

func TestCaptureInactiveHold(t *testing.T) {
	svc, _ := newService()
	w, _ := svc.Create("@user:alice", asset)
	svc.Credit(w.ID, "", 100000, "")
	h, _ := svc.Hold(w.ID, "", 30000, "", "")
	svc.CaptureHold(w.ID, h.ID, "")

	_, err := svc.CaptureHold(w.ID, h.ID, "") // already captured
	we, ok := err.(*Error)
	if !ok || we.Code != ErrHoldNotActive {
		t.Fatalf("expected ERR_HOLD_NOT_ACTIVE, got %v", err)
	}
}

func TestWalletNotFound(t *testing.T) {
	svc, _ := newService()
	if _, err := svc.Balances("wlt_does_not_exist"); err == nil {
		t.Fatal("expected ERR_NOT_FOUND")
	}
}

func TestTransfer(t *testing.T) {
	svc, _ := newService()
	alice, _ := svc.Create("@user:alice", asset)
	bob, _ := svc.Create("@user:bob", asset)
	svc.Credit(alice.ID, "", 100000, "")

	// nominal
	if _, err := svc.Transfer(alice.ID, bob.ID, "", 30000); err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if av := mustBal(t, svc, alice.ID); av != 70000 {
		t.Fatalf("alice available expected 70000, got %d", av)
	}
	if av := mustBal(t, svc, bob.ID); av != 30000 {
		t.Fatalf("bob available expected 30000, got %d", av)
	}

	// destination wallet doesn't exist
	if _, err := svc.Transfer(alice.ID, "wlt_ghost", "", 1000); err == nil {
		t.Fatal("expected ERR_NOT_FOUND for missing destination wallet")
	}

	// insufficient funds (no implicit debt = no riba)
	_, err := svc.Transfer(alice.ID, bob.ID, "", 999999)
	we, ok := err.(*Error)
	if !ok || we.Code != ErrInsufficientFunds {
		t.Fatalf("expected ERR_INSUFFICIENT_FUNDS, got %v", err)
	}
}

func TestReleaseHoldReturnsFunds(t *testing.T) {
	svc, _ := newService()
	w, _ := svc.Create("@user:alice", asset)
	svc.Credit(w.ID, "", 50000, "")
	h, _ := svc.Hold(w.ID, "", 20000, "escrow", "2026-12-31T00:00:00Z")
	if h.ExpiresAt != "2026-12-31T00:00:00Z" || h.Reason != "escrow" {
		t.Fatalf("hold should record reason+expires_at, got %+v", h)
	}
	if _, err := svc.Release(w.ID, h.ID); err != nil {
		t.Fatalf("release: %v", err)
	}
	b, _ := svc.Balances(w.ID)
	if b.Available[asset] != 50000 || b.Held[asset] != 0 {
		t.Fatalf("after release expected available 50000 / held 0, got %d / %d", b.Available[asset], b.Held[asset])
	}
}

func mustBal(t *testing.T, svc *Service, id string) int64 {
	t.Helper()
	b, err := svc.Balances(id)
	if err != nil {
		t.Fatalf("balances: %v", err)
	}
	return b.Available[asset]
}

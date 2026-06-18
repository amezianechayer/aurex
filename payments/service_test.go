package payments

import (
	"fmt"
	"strings"
	"testing"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/wallets"
)

// --- fakes ---

// fakeLedger applies postings with the same non-negative invariant as the real
// ledger (only @world may go negative), so payin/payout/compensation behave for
// real.
type fakeLedger struct {
	bal           map[string]int64
	failRefPrefix string // if set, Commit fails for any tx whose Reference starts with it
}

func newFakeLedger() *fakeLedger { return &fakeLedger{bal: map[string]int64{}} }

func acctKey(a, asset string) string { return a + "|" + asset }

func (f *fakeLedger) Commit(ts []core.Transaction) ([]core.Transaction, error) {
	if f.failRefPrefix != "" {
		for _, t := range ts {
			if strings.HasPrefix(t.Reference, f.failRefPrefix) {
				return nil, fmt.Errorf("ledger commit blocked for %q", t.Reference)
			}
		}
	}
	work := map[string]int64{}
	for k, v := range f.bal {
		work[k] = v
	}
	for _, t := range ts {
		for _, p := range t.Postings {
			if p.Source != core.WORLD {
				if work[acctKey(p.Source, p.Asset)]-p.Amount < 0 {
					return nil, fmt.Errorf("balance.insufficient.%s", p.Asset)
				}
				work[acctKey(p.Source, p.Asset)] -= p.Amount
			}
			if p.Destination != core.WORLD {
				work[acctKey(p.Destination, p.Asset)] += p.Amount
			}
		}
	}
	f.bal = work
	return ts, nil
}

func (f *fakeLedger) setWallet(walletID, asset string, amount int64) {
	f.bal[acctKey(wallets.MainAccount(walletID), asset)] = amount
}
func (f *fakeLedger) walletBal(walletID, asset string) int64 {
	return f.bal[acctKey(wallets.MainAccount(walletID), asset)]
}

type fakeStore struct {
	rows map[string]Payment
	seen map[string]bool // (psp|external_id) for non-empty external_id — mirrors the unique index
}

func newFakeStore() *fakeStore {
	return &fakeStore{rows: map[string]Payment{}, seen: map[string]bool{}}
}
func (s *fakeStore) SavePayment(p Payment) error {
	if p.ExternalID != "" {
		k := p.PSP + "|" + p.ExternalID
		if s.seen[k] {
			return fmt.Errorf("UNIQUE constraint failed: payments.psp, payments.external_id")
		}
		s.seen[k] = true
	}
	s.rows[p.ID] = p
	return nil
}
func (s *fakeStore) UpdatePayment(p Payment) error { s.rows[p.ID] = p; return nil }
func (s *fakeStore) GetPaymentByExternalID(psp, externalID string) (Payment, error) {
	for _, p := range s.rows {
		if p.PSP == psp && p.ExternalID == externalID {
			return p, nil
		}
	}
	return Payment{}, fmt.Errorf("not found")
}
func (s *fakeStore) GetPayment(id string) (Payment, error) {
	p, ok := s.rows[id]
	if !ok {
		return Payment{}, fmt.Errorf("not found")
	}
	return p, nil
}
func (s *fakeStore) ListPayments(limit, offset int) ([]Payment, error) {
	out := []Payment{}
	for _, p := range s.rows {
		out = append(out, p)
	}
	return out, nil
}

// fakeConn lets each test preset the webhook event / payout outcome.
type fakeConn struct {
	event       Event
	eventErr    error
	payout      Payout
	payoutErr   error
	payoutCalls int
}

func (c *fakeConn) Name() string { return "fake" }
func (c *fakeConn) CreatePayment(amount int64, asset, ref string) (Payment, error) {
	return Payment{ExternalID: "pi_fake", Status: StatusPending, Amount: amount, Asset: asset, Reference: ref}, nil
}
func (c *fakeConn) CreatePayout(amount int64, asset, dest, ref string) (Payout, error) {
	c.payoutCalls++
	return c.payout, c.payoutErr
}
func (c *fakeConn) HandleWebhook(payload []byte, signature string) (Event, error) {
	return c.event, c.eventErr
}

func newSvc(conn PSP) (*Service, *fakeLedger, *fakeStore) {
	led, st := newFakeLedger(), newFakeStore()
	return NewService(led, st, NewRegistry(conn)), led, st
}

const asset = "AED.2"

// --- webhook (payin) ---

func TestWebhookCreditsWallet(t *testing.T) {
	conn := &fakeConn{event: Event{Type: EventPaymentSucceeded, PSP: "fake", WalletID: "w1", Asset: asset, Amount: 100000, ExternalID: "pi_1"}}
	svc, led, st := newSvc(conn)

	ev, err := svc.HandleWebhook("fake", []byte("{}"), "sig")
	if err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if ev.Type != EventPaymentSucceeded {
		t.Fatalf("expected succeeded event, got %v", ev.Type)
	}
	if led.walletBal("w1", asset) != 100000 {
		t.Fatalf("expected wallet credited 100000, got %d", led.walletBal("w1", asset))
	}
	// the @psp clearing account nets to zero
	if led.bal[acctKey(PSPAccount("fake"), asset)] != 0 {
		t.Fatalf("expected @psp:fake to net to zero, got %d", led.bal[acctKey(PSPAccount("fake"), asset)])
	}
	if len(st.rows) != 1 {
		t.Fatalf("expected 1 payment recorded, got %d", len(st.rows))
	}
}

func TestWebhookUnknownPSP(t *testing.T) {
	svc, _, _ := newSvc(&fakeConn{})
	_, err := svc.HandleWebhook("nope", []byte("{}"), "sig")
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrUnknownPSP {
		t.Fatalf("expected ERR_UNKNOWN_PSP, got %v", err)
	}
}

func TestWebhookFailedEventDoesNotCredit(t *testing.T) {
	conn := &fakeConn{event: Event{Type: EventPaymentFailed, WalletID: "w1", Asset: asset, Amount: 100000}}
	svc, led, st := newSvc(conn)
	if _, err := svc.HandleWebhook("fake", []byte("{}"), "sig"); err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if led.walletBal("w1", asset) != 0 || len(st.rows) != 0 {
		t.Fatal("a failed event must not credit nor record a settled payin")
	}
}

func TestWebhookBadSignatureBubblesUp(t *testing.T) {
	conn := &fakeConn{eventErr: newError(ErrInvalidSignature, "bad sig")}
	svc, led, _ := newSvc(conn)
	_, err := svc.HandleWebhook("fake", []byte("{}"), "bad")
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrInvalidSignature {
		t.Fatalf("expected ERR_INVALID_SIGNATURE, got %v", err)
	}
	if led.walletBal("w1", asset) != 0 {
		t.Fatal("nothing should be credited on a bad signature")
	}
}

// --- payout ---

func TestPayoutSuccess(t *testing.T) {
	conn := &fakeConn{payout: Payout{ExternalID: "po_ext", Status: StatusSucceeded}}
	svc, led, st := newSvc(conn)
	led.setWallet("w1", asset, 100000)

	rec, err := svc.CreatePayout("fake", "w1", asset, "acct_bank", 30000)
	if err != nil {
		t.Fatalf("payout: %v", err)
	}
	if rec.Status != StatusSucceeded || rec.ExternalID != "po_ext" {
		t.Fatalf("expected succeeded payout with external id, got %+v", rec)
	}
	if led.walletBal("w1", asset) != 70000 {
		t.Fatalf("expected wallet debited to 70000, got %d", led.walletBal("w1", asset))
	}
	if st.rows[rec.ID].Status != StatusSucceeded {
		t.Fatal("payment record should be succeeded")
	}
}

func TestPayoutPSPErrorCompensates(t *testing.T) {
	conn := &fakeConn{payoutErr: fmt.Errorf("stripe 500")}
	svc, led, st := newSvc(conn)
	led.setWallet("w1", asset, 100000)

	rec, err := svc.CreatePayout("fake", "w1", asset, "acct_bank", 30000)
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrPSPUnavailable {
		t.Fatalf("expected ERR_PSP_UNAVAILABLE, got %v", err)
	}
	if led.walletBal("w1", asset) != 100000 {
		t.Fatalf("expected funds compensated back to 100000, got %d", led.walletBal("w1", asset))
	}
	if st.rows[rec.ID].Status != StatusFailed {
		t.Fatalf("expected payment marked failed, got %s", st.rows[rec.ID].Status)
	}
}

func TestPayoutInsufficientFundsNoPSPCall(t *testing.T) {
	conn := &fakeConn{payout: Payout{ExternalID: "po_ext"}}
	svc, led, _ := newSvc(conn)
	led.setWallet("w1", asset, 5000) // less than the payout

	_, err := svc.CreatePayout("fake", "w1", asset, "acct_bank", 30000)
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrInsufficientFunds {
		t.Fatalf("expected ERR_INSUFFICIENT_FUNDS, got %v", err)
	}
	if conn.payoutCalls != 0 {
		t.Fatal("PSP must NOT be called when the wallet debit fails (no money to pay out)")
	}
}

func TestPayoutUnknownPSP(t *testing.T) {
	svc, _, _ := newSvc(&fakeConn{})
	_, err := svc.CreatePayout("nope", "w1", asset, "acct", 1000)
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrUnknownPSP {
		t.Fatalf("expected ERR_UNKNOWN_PSP, got %v", err)
	}
}

// C1: a redelivered webhook (same external_id) credits the wallet exactly once.
func TestWebhookReplayCreditsOnce(t *testing.T) {
	conn := &fakeConn{event: Event{Type: EventPaymentSucceeded, PSP: "fake", WalletID: "w1", Asset: asset, Amount: 100000, ExternalID: "pi_dup"}}
	svc, led, st := newSvc(conn)

	if _, err := svc.HandleWebhook("fake", []byte("{}"), "sig"); err != nil {
		t.Fatalf("first delivery: %v", err)
	}
	// PSP redelivers the identical event
	if _, err := svc.HandleWebhook("fake", []byte("{}"), "sig"); err != nil {
		t.Fatalf("replay should be an idempotent no-op, got: %v", err)
	}
	if led.walletBal("w1", asset) != 100000 {
		t.Fatalf("expected wallet credited exactly once (100000), got %d", led.walletBal("w1", asset))
	}
	if len(st.rows) != 1 {
		t.Fatalf("expected exactly one payment row, got %d", len(st.rows))
	}
}

// --- payin (top-up) initiation ---

// CreatePayin records a pending payin and does NOT move money: the external
// funds are not in yet, so crediting now would be fictitious.
func TestPayinInitiatesPendingWithoutCrediting(t *testing.T) {
	conn := &fakeConn{}
	svc, led, st := newSvc(conn)

	rec, err := svc.CreatePayin("fake", "w1", asset, "", 100000)
	if err != nil {
		t.Fatalf("payin: %v", err)
	}
	if rec.Direction != DirectionPayin || rec.Status != StatusPending {
		t.Fatalf("expected a pending payin, got %+v", rec)
	}
	if rec.ExternalID != "pi_fake" {
		t.Fatalf("expected the PSP external id on the record, got %q", rec.ExternalID)
	}
	if led.walletBal("w1", asset) != 0 {
		t.Fatalf("no money must move at initiation, wallet=%d", led.walletBal("w1", asset))
	}
	if len(st.rows) != 1 || st.rows[rec.ID].Status != StatusPending {
		t.Fatal("expected exactly one pending payment row")
	}
}

// The webhook reconciles the pending initiation: it credits the wallet once and
// flips the record to succeeded. A redelivery is an idempotent no-op.
func TestPayinSettledByWebhookExactlyOnce(t *testing.T) {
	conn := &fakeConn{event: Event{Type: EventPaymentSucceeded, PSP: "fake", WalletID: "w1", Asset: asset, Amount: 100000, ExternalID: "pi_fake"}}
	svc, led, st := newSvc(conn)

	rec, err := svc.CreatePayin("fake", "w1", asset, "", 100000)
	if err != nil {
		t.Fatalf("payin: %v", err)
	}
	if _, err := svc.HandleWebhook("fake", []byte("{}"), "sig"); err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if led.walletBal("w1", asset) != 100000 {
		t.Fatalf("expected wallet credited 100000, got %d", led.walletBal("w1", asset))
	}
	if st.rows[rec.ID].Status != StatusSucceeded {
		t.Fatalf("expected the initiated payin marked succeeded, got %s", st.rows[rec.ID].Status)
	}
	if len(st.rows) != 1 {
		t.Fatalf("expected exactly one payment row (the initiation, settled), got %d", len(st.rows))
	}
	// redelivery: no second credit, no new row
	if _, err := svc.HandleWebhook("fake", []byte("{}"), "sig"); err != nil {
		t.Fatalf("redelivery should be a no-op, got %v", err)
	}
	if led.walletBal("w1", asset) != 100000 || len(st.rows) != 1 {
		t.Fatalf("redelivery must not credit again: bal=%d rows=%d", led.walletBal("w1", asset), len(st.rows))
	}
}

func TestPayinUnknownPSP(t *testing.T) {
	svc, _, _ := newSvc(&fakeConn{})
	_, err := svc.CreatePayin("nope", "w1", asset, "", 1000)
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrUnknownPSP {
		t.Fatalf("expected ERR_UNKNOWN_PSP, got %v", err)
	}
}

func TestPayinInvalidAmount(t *testing.T) {
	svc, _, _ := newSvc(&fakeConn{})
	_, err := svc.CreatePayin("fake", "w1", asset, "", 0)
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrInvalidParams {
		t.Fatalf("expected ERR_INVALID_PARAMS, got %v", err)
	}
}

// C2: if the PSP fails AND the compensation reversal also fails, the wallet is
// left debited and the caller gets ERR_COMPENSATION_FAILED (not a silent loss).
func TestPayoutCompensationFailure(t *testing.T) {
	conn := &fakeConn{payoutErr: fmt.Errorf("stripe 500")}
	svc, led, _ := newSvc(conn)
	led.setWallet("w1", asset, 100000)
	led.failRefPrefix = "compensate:" // the reversal posting will be rejected

	rec, err := svc.CreatePayout("fake", "w1", asset, "acct_bank", 30000)
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrCompensationFailed {
		t.Fatalf("expected ERR_COMPENSATION_FAILED, got %v", err)
	}
	// funds remain debited (the whole point of the alert): 100000 - 30000
	if led.walletBal("w1", asset) != 70000 {
		t.Fatalf("expected wallet still debited at 70000 after failed compensation, got %d", led.walletBal("w1", asset))
	}
	if rec.Status != StatusFailed {
		t.Fatalf("expected payment marked failed, got %s", rec.Status)
	}
}

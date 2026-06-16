package payments

import (
	"strings"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/wallets"
)

// Service orchestrates PSP flows on top of the ledger + wallets. It holds no
// balances: a payin and a payout are each a single, atomic two-leg ledger
// transaction routed through the Guard. The @psp:<name> clearing account nets
// to zero; the wallet's main account carries the real balance.
//
//	payin  : @world -> @psp:<name> -> @wallets:<id>:main      (external money in)
//	payout : @wallets:<id>:main -> @psp:<name> -> @world      (debit first; the
//	         ledger refuses it if the wallet is short — no payout you can't fund)
type Service struct {
	led   LedgerPort
	store PaymentStore
	reg   *Registry
}

func NewService(led LedgerPort, store PaymentStore, reg *Registry) *Service {
	return &Service{led: led, store: store, reg: reg}
}

func payinPostings(psp, walletID, asset string, amount int64) []core.Posting {
	return []core.Posting{
		{Source: core.WORLD, Destination: PSPAccount(psp), Asset: asset, Amount: amount},
		{Source: PSPAccount(psp), Destination: wallets.MainAccount(walletID), Asset: asset, Amount: amount},
	}
}

func payoutPostings(psp, walletID, asset string, amount int64) []core.Posting {
	return []core.Posting{
		{Source: wallets.MainAccount(walletID), Destination: PSPAccount(psp), Asset: asset, Amount: amount},
		{Source: PSPAccount(psp), Destination: core.WORLD, Asset: asset, Amount: amount},
	}
}

func (s *Service) commit(postings []core.Posting, ref string) error {
	_, err := s.led.Commit([]core.Transaction{{Postings: postings, Reference: ref}})
	return mapCommitErr(err)
}

// HandleWebhook validates and parses a provider webhook, and on a successful
// payment credits the target wallet and records the payin.
func (s *Service) HandleWebhook(pspName string, payload []byte, signature string) (Event, error) {
	conn, ok := s.reg.Get(pspName)
	if !ok {
		return Event{}, &Error{Code: ErrUnknownPSP, Message: "unknown psp: " + pspName, PSP: pspName}
	}
	ev, err := conn.HandleWebhook(payload, signature)
	if err != nil {
		return Event{}, err // invalid signature / payload, already typed
	}
	if ev.Type != EventPaymentSucceeded {
		return ev, nil // nothing to settle — acknowledged
	}
	if ev.WalletID == "" {
		return Event{}, newError(ErrInvalidPayload, "event carries no wallet_id")
	}
	if err := s.commit(payinPostings(pspName, ev.WalletID, ev.Asset, ev.Amount), generateID("payin_")); err != nil {
		return Event{}, err // wallet/guard error surfaced to the caller
	}
	ts := now()
	_ = s.store.SavePayment(Payment{
		ID: generateID("pay_"), PSP: pspName, Direction: DirectionPayin,
		WalletID: ev.WalletID, Asset: ev.Asset, Amount: ev.Amount,
		Status: StatusSucceeded, Reference: ev.Reference, ExternalID: ev.ExternalID,
		CreatedAt: ts, UpdatedAt: ts,
	})
	return ev, nil
}

// CreatePayout debits the wallet FIRST (the ledger refuses it if the balance is
// short), then initiates the payout at the PSP. If the PSP call fails, the debit
// is compensated (funds credited back) and the payout is marked failed.
func (s *Service) CreatePayout(pspName, walletID, asset, destination string, amount int64) (Payment, error) {
	conn, ok := s.reg.Get(pspName)
	if !ok {
		return Payment{}, &Error{Code: ErrUnknownPSP, Message: "unknown psp: " + pspName, PSP: pspName}
	}
	if amount <= 0 {
		return Payment{}, newError(ErrInvalidParams, "amount must be > 0")
	}
	if walletID == "" || asset == "" || destination == "" {
		return Payment{}, newError(ErrInvalidParams, "wallet_id, asset and destination are required")
	}

	ref := generateID("po_")
	// 1. debit the wallet out through the PSP clearing account to @world
	if err := s.commit(payoutPostings(pspName, walletID, asset, amount), ref); err != nil {
		return Payment{}, err
	}
	ts := now()
	rec := Payment{
		ID: ref, PSP: pspName, Direction: DirectionPayout, WalletID: walletID,
		Asset: asset, Amount: amount, Status: StatusPending, Reference: ref,
		CreatedAt: ts, UpdatedAt: ts,
	}
	_ = s.store.SavePayment(rec)

	// 2. initiate at the PSP
	po, err := conn.CreatePayout(amount, asset, destination, ref)
	if err != nil {
		// compensate: the funds never reached the bank, so reverse them back in
		_ = s.commit(payinPostings(pspName, walletID, asset, amount), "compensate:"+ref)
		rec.Status = StatusFailed
		rec.UpdatedAt = now()
		_ = s.store.UpdatePayment(rec)
		return rec, &Error{Code: ErrPSPUnavailable, Message: "payout failed at psp: " + err.Error(), PSP: pspName}
	}

	rec.Status = StatusSucceeded
	rec.ExternalID = po.ExternalID
	rec.UpdatedAt = now()
	_ = s.store.UpdatePayment(rec)
	return rec, nil
}

func (s *Service) ListPayments(limit, offset int) ([]Payment, error) {
	return s.store.ListPayments(limit, offset)
}

// mapCommitErr turns the ledger's insufficient-balance error into a typed
// payments error; other Commit errors (e.g. a *guard.Error) pass through.
func mapCommitErr(err error) error {
	if err == nil {
		return nil
	}
	if strings.HasPrefix(err.Error(), "balance.insufficient") {
		return newError(ErrInsufficientFunds, "insufficient wallet balance")
	}
	return err
}

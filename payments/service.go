package payments

import (
	"log"
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

	// Reconcile with a payin that was already initiated (recorded as pending by
	// CreatePayin): settle it once. This is also the redelivery gate — a second
	// webhook for an already-succeeded payment is an idempotent no-op (C1).
	if ev.ExternalID != "" {
		if existing, err := s.store.GetPaymentByExternalID(pspName, ev.ExternalID); err == nil {
			if existing.Status == StatusSucceeded {
				return ev, nil // already settled — idempotent no-op
			}
			if cerr := s.commit(payinPostings(pspName, ev.WalletID, ev.Asset, ev.Amount), generateID("payin_")); cerr != nil {
				existing.Status = StatusFailed
				existing.UpdatedAt = now()
				_ = s.store.UpdatePayment(existing)
				return Event{}, cerr
			}
			existing.Status = StatusSucceeded
			existing.UpdatedAt = now()
			_ = s.store.UpdatePayment(existing)
			return ev, nil
		}
	}

	// No prior record (webhook-only payin, no preceding initiation). Reserve
	// (psp, external_id) by writing the record BEFORE crediting — a concurrent
	// redelivery collides on the unique index and is acknowledged without a
	// second credit. A genuine store error (not a duplicate) aborts before any
	// money moves.
	ts := now()
	rec := Payment{
		ID: generateID("pay_"), PSP: pspName, Direction: DirectionPayin,
		WalletID: ev.WalletID, Asset: ev.Asset, Amount: ev.Amount,
		Status: StatusPending, Reference: ev.Reference, ExternalID: ev.ExternalID,
		CreatedAt: ts, UpdatedAt: ts,
	}
	if ev.ExternalID != "" {
		if err := s.store.SavePayment(rec); err != nil {
			if isDuplicateErr(err) {
				return ev, nil // a concurrent delivery won the reservation
			}
			return Event{}, newError(ErrInternal, "cannot record payin: "+err.Error())
		}
	}
	if err := s.commit(payinPostings(pspName, ev.WalletID, ev.Asset, ev.Amount), generateID("payin_")); err != nil {
		if ev.ExternalID != "" {
			rec.Status = StatusFailed
			rec.UpdatedAt = now()
			_ = s.store.UpdatePayment(rec)
		}
		return Event{}, err // wallet/guard error surfaced to the caller
	}
	rec.Status = StatusSucceeded
	rec.UpdatedAt = now()
	if ev.ExternalID != "" {
		_ = s.store.UpdatePayment(rec)
	} else {
		_ = s.store.SavePayment(rec)
	}
	return ev, nil
}

// CreatePayin initiates an inbound payment (a top-up) at the PSP and records it
// as pending. NO ledger posting happens here: the external money is not in yet,
// so crediting now would be fictitious. The wallet is credited when the PSP
// confirms the payment via its webhook (HandleWebhook reconciles by external_id).
// The returned record carries the PSP external id the client uses to complete
// the charge.
func (s *Service) CreatePayin(pspName, walletID, asset, reference string, amount int64) (Payment, error) {
	conn, ok := s.reg.Get(pspName)
	if !ok {
		return Payment{}, &Error{Code: ErrUnknownPSP, Message: "unknown psp: " + pspName, PSP: pspName}
	}
	if amount <= 0 {
		return Payment{}, newError(ErrInvalidParams, "amount must be > 0")
	}
	if walletID == "" || asset == "" {
		return Payment{}, newError(ErrInvalidParams, "wallet_id and asset are required")
	}

	ref := reference
	if ref == "" {
		ref = generateID("pi_")
	}

	// initiate at the PSP (e.g. a Stripe PaymentIntent); no funds move yet
	po, err := conn.CreatePayment(amount, asset, ref)
	if err != nil {
		return Payment{}, &Error{Code: ErrPSPUnavailable, Message: "payin initiation failed at psp: " + err.Error(), PSP: pspName}
	}

	ts := now()
	rec := Payment{
		ID: generateID("pay_"), PSP: pspName, Direction: DirectionPayin,
		WalletID: walletID, Asset: asset, Amount: amount,
		Status: StatusPending, Reference: ref, ExternalID: po.ExternalID,
		CreatedAt: ts, UpdatedAt: ts,
	}
	if err := s.store.SavePayment(rec); err != nil {
		if isDuplicateErr(err) {
			return Payment{}, &Error{Code: ErrDuplicate, Message: "a payment with this external id already exists", PSP: pspName}
		}
		return Payment{}, newError(ErrInternal, "cannot record payin: "+err.Error())
	}
	return rec, nil
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
	// Can't record the payout → reverse the debit and abort BEFORE calling the
	// PSP, so we never pay out an unrecorded transfer (M1).
	if err := s.store.SavePayment(rec); err != nil {
		return Payment{}, s.reverse(pspName, walletID, asset, amount, ref, "cannot record payout", err)
	}

	// 2. initiate at the PSP
	po, err := conn.CreatePayout(amount, asset, destination, ref)
	if err != nil {
		// compensate: the funds never reached the bank, reverse them back in. If
		// the reversal ALSO fails the wallet is debited with nothing paid out —
		// surface a distinct error so monitoring can alert for manual recon (C2).
		if cerr := s.commit(payinPostings(pspName, walletID, asset, amount), "compensate:"+ref); cerr != nil {
			log.Printf("CRITICAL payments: payout failed AND compensation failed psp=%s wallet=%s amount=%d ref=%s psp_err=%v compensate_err=%v",
				pspName, walletID, amount, ref, err, cerr)
			rec.Status = StatusFailed
			rec.UpdatedAt = now()
			_ = s.store.UpdatePayment(rec)
			return rec, &Error{Code: ErrCompensationFailed, Message: "payout failed and compensation failed — funds debited, manual reconciliation required: " + cerr.Error(), PSP: pspName}
		}
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

// reverse credits a just-committed payout debit back to the wallet. If the
// reversal itself fails, the funds are stranded debited — return the distinct
// ErrCompensationFailed and log for manual reconciliation.
func (s *Service) reverse(pspName, walletID, asset string, amount int64, ref, why string, cause error) error {
	if cerr := s.commit(payinPostings(pspName, walletID, asset, amount), "reverse:"+ref); cerr != nil {
		log.Printf("CRITICAL payments: %s AND reversal failed psp=%s wallet=%s amount=%d ref=%s cause=%v reverse_err=%v",
			why, pspName, walletID, amount, ref, cause, cerr)
		return &Error{Code: ErrCompensationFailed, Message: why + " and reversal failed — funds debited, manual reconciliation required: " + cerr.Error(), PSP: pspName}
	}
	return newError(ErrInternal, why+", debit reversed: "+cause.Error())
}

func (s *Service) ListPayments(limit, offset int) ([]Payment, error) {
	return s.store.ListPayments(limit, offset)
}

// isDuplicateErr reports whether a store error is a unique-constraint violation
// (used as the webhook idempotency gate).
func isDuplicateErr(err error) bool {
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unique") || strings.Contains(s, "duplicate")
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

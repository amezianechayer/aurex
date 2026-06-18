package payments

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func stripeEvent(t *testing.T, walletID string, amount int64) []byte {
	t.Helper()
	ev := map[string]interface{}{
		"type": "payment_intent.succeeded",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":       "pi_123",
				"amount":   amount,
				"currency": "aed",
				"metadata": map[string]string{"wallet_id": walletID, "asset": "AED.2", "reference": "topup-1"},
			},
		},
	}
	b, _ := json.Marshal(ev)
	return b
}

func TestStripeWebhookValidSignature(t *testing.T) {
	s := NewStripe(StripeConfig{WebhookSecret: "whsec_test"})
	payload := stripeEvent(t, "wlt_alice", 100000)
	sig := s.SignWebhook(payload, "1700000000")

	ev, err := s.HandleWebhook(payload, sig)
	if err != nil {
		t.Fatalf("valid webhook: %v", err)
	}
	if ev.Type != EventPaymentSucceeded || ev.WalletID != "wlt_alice" || ev.Amount != 100000 || ev.Asset != "AED.2" {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestStripeWebhookBadSignature(t *testing.T) {
	s := NewStripe(StripeConfig{WebhookSecret: "whsec_test"})
	payload := stripeEvent(t, "wlt_alice", 100000)
	_, err := s.HandleWebhook(payload, "t=1700000000,v1=deadbeef")
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrInvalidSignature {
		t.Fatalf("expected ERR_INVALID_SIGNATURE, got %v", err)
	}
}

func TestStripeCreatePayoutSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/payouts" {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"po_live_123","object":"payout"}`))
	}))
	defer srv.Close()

	s := NewStripe(StripeConfig{SecretKey: "sk_test", BaseURL: srv.URL, HTTP: srv.Client()})
	po, err := s.CreatePayout(50000, "AED.2", "acct_bank", "po_ref")
	if err != nil {
		t.Fatalf("payout: %v", err)
	}
	if po.ExternalID != "po_live_123" {
		t.Fatalf("expected external id po_live_123, got %q", po.ExternalID)
	}
}

func TestStripeCreatePaymentReturnsClientSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/payment_intents" {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"pi_live_1","client_secret":"pi_live_1_secret_abc","object":"payment_intent"}`))
	}))
	defer srv.Close()

	s := NewStripe(StripeConfig{SecretKey: "sk_test", BaseURL: srv.URL, HTTP: srv.Client()})
	p, err := s.CreatePayment(100000, "AED.2", "pi_ref")
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if p.ExternalID != "pi_live_1" || p.ClientSecret != "pi_live_1_secret_abc" {
		t.Fatalf("expected external id + client secret, got %+v", p)
	}
	if p.Status != StatusPending {
		t.Fatalf("expected pending status, got %s", p.Status)
	}
}

func TestStripeCreatePayoutError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"boom"}}`))
	}))
	defer srv.Close()

	s := NewStripe(StripeConfig{SecretKey: "sk_test", BaseURL: srv.URL, HTTP: srv.Client()})
	_, err := s.CreatePayout(50000, "AED.2", "acct_bank", "po_ref")
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrPSPUnavailable {
		t.Fatalf("expected ERR_PSP_UNAVAILABLE, got %v", err)
	}
}

// Real Stripe connector (genuine HMAC verification) wired through the Service:
// the closest "simulated webhook" to production — a correctly-signed payload
// credits the wallet.
func TestStripeWebhookThroughService(t *testing.T) {
	s := NewStripe(StripeConfig{WebhookSecret: "whsec_test"})
	svc := NewService(newFakeLedger(), newFakeStore(), NewRegistry(s))

	payload := stripeEvent(t, "wlt_bob", 75000)
	sig := s.SignWebhook(payload, "1700000000")

	led := svc.led.(*fakeLedger)
	if _, err := svc.HandleWebhook("stripe", payload, sig); err != nil {
		t.Fatalf("service webhook: %v", err)
	}
	if led.walletBal("wlt_bob", "AED.2") != 75000 {
		t.Fatalf("expected wallet credited 75000, got %d", led.walletBal("wlt_bob", "AED.2"))
	}
}

func paytabsCallback(status string) []byte {
	cb := map[string]interface{}{
		"tran_ref":      "TST2200001",
		"cart_id":       "topup-7",
		"cart_amount":   "300.00",
		"cart_currency": "AED",
		"user_defined":  map[string]string{"udf1": "wlt_alice", "udf2": "AED.2"},
		"payment_result": map[string]string{"response_status": status},
	}
	b, _ := json.Marshal(cb)
	return b
}

func TestPayTabsWebhookValidSignature(t *testing.T) {
	p := NewPayTabs(PayTabsConfig{ServerKey: "SJKEY"})
	payload := paytabsCallback("A")
	sig := p.SignWebhook(payload)

	ev, err := p.HandleWebhook(payload, sig)
	if err != nil {
		t.Fatalf("valid webhook: %v", err)
	}
	if ev.Type != EventPaymentSucceeded || ev.WalletID != "wlt_alice" || ev.Asset != "AED.2" || ev.Amount != 30000 {
		t.Fatalf("unexpected paytabs event: %+v", ev)
	}
}

func TestPayTabsCreatePaymentReturnsRedirectURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payment/request" {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tran_ref":"TST22","redirect_url":"https://secure.paytabs.com/payment/page/abc"}`))
	}))
	defer srv.Close()

	p := NewPayTabs(PayTabsConfig{ServerKey: "SJKEY", ProfileID: "123", BaseURL: srv.URL, HTTP: srv.Client()})
	pay, err := p.CreatePayment(30000, "AED.2", "topup-9")
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if pay.ExternalID != "TST22" || pay.RedirectURL != "https://secure.paytabs.com/payment/page/abc" {
		t.Fatalf("expected tran_ref + redirect_url, got %+v", pay)
	}
}

func TestPayTabsWebhookBadSignature(t *testing.T) {
	p := NewPayTabs(PayTabsConfig{ServerKey: "SJKEY"})
	_, err := p.HandleWebhook(paytabsCallback("A"), "deadbeef")
	pe, ok := err.(*Error)
	if !ok || pe.Code != ErrInvalidSignature {
		t.Fatalf("expected ERR_INVALID_SIGNATURE, got %v", err)
	}
}

func TestParseMinor(t *testing.T) {
	cases := []struct {
		in  string
		out int64
	}{{"300.00", 30000}, {"10", 1000}, {"10.5", 1050}, {"0.99", 99}, {"10.567", 1056}, {"", 0}}
	for _, c := range cases {
		if got := parseMinor(c.in, 2); got != c.out {
			t.Fatalf("parseMinor(%q) = %d, want %d", c.in, got, c.out)
		}
	}
	if formatMinor(30000, 2) != "300.00" {
		t.Fatalf("formatMinor(30000) = %q", formatMinor(30000, 2))
	}
}

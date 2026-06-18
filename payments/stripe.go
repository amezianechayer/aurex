package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// StripeConfig configures the Stripe connector. BaseURL and HTTP are injectable
// so the connector can be tested against an httptest server without live keys.
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
	BaseURL       string
	HTTP          *http.Client
}

type Stripe struct{ cfg StripeConfig }

func NewStripe(cfg StripeConfig) *Stripe {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.stripe.com"
	}
	if cfg.HTTP == nil {
		cfg.HTTP = http.DefaultClient
	}
	return &Stripe{cfg: cfg}
}

func (s *Stripe) Name() string { return "stripe" }

// HandleWebhook verifies the Stripe-Signature header (HMAC-SHA256 over
// "timestamp.payload") and normalizes the event. wallet_id and (optionally)
// asset travel in the object metadata, set when the payment was created.
func (s *Stripe) HandleWebhook(payload []byte, signature string) (Event, error) {
	if err := s.verifySignature(payload, signature); err != nil {
		return Event{}, err
	}
	var ev struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID       string            `json:"id"`
				Amount   int64             `json:"amount"`
				Currency string            `json:"currency"`
				Metadata map[string]string `json:"metadata"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &ev); err != nil {
		return Event{}, newError(ErrInvalidPayload, "invalid stripe event json")
	}
	obj := ev.Data.Object
	asset := obj.Metadata["asset"]
	if asset == "" {
		asset = strings.ToUpper(obj.Currency) + ".2"
	}
	out := Event{
		Type: EventUnhandled, PSP: s.Name(), ExternalID: obj.ID,
		WalletID: obj.Metadata["wallet_id"], Reference: obj.Metadata["reference"],
		Asset: asset, Amount: obj.Amount,
	}
	switch ev.Type {
	case "payment_intent.succeeded", "checkout.session.completed", "charge.succeeded":
		out.Type = EventPaymentSucceeded
	case "payment_intent.payment_failed", "charge.failed":
		out.Type = EventPaymentFailed
	}
	return out, nil
}

func (s *Stripe) verifySignature(payload []byte, header string) error {
	if s.cfg.WebhookSecret == "" {
		return newError(ErrInvalidSignature, "no webhook secret configured")
	}
	var t, v1 string
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			t = kv[1]
		case "v1":
			if v1 == "" {
				v1 = kv[1]
			}
		}
	}
	if t == "" || v1 == "" {
		return newError(ErrInvalidSignature, "malformed stripe-signature header")
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.WebhookSecret))
	mac.Write([]byte(t + "." + string(payload)))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(v1)) {
		return newError(ErrInvalidSignature, "stripe signature mismatch")
	}
	return nil
}

// SignWebhook builds a valid Stripe-Signature header for the payload. Exposed so
// callers (and tests) can produce signed payloads deterministically.
func (s *Stripe) SignWebhook(payload []byte, timestamp string) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.WebhookSecret))
	mac.Write([]byte(timestamp + "." + string(payload)))
	return "t=" + timestamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func (s *Stripe) CreatePayment(amount int64, asset, ref string) (Payment, error) {
	form := url.Values{}
	form.Set("amount", strconv.FormatInt(amount, 10))
	form.Set("currency", currencyOf(asset))
	form.Set("metadata[reference]", ref)
	res, err := s.post("/v1/payment_intents", form)
	if err != nil {
		return Payment{}, err
	}
	return Payment{
		ExternalID: stringField(res, "id"), Status: StatusPending,
		Amount: amount, Asset: asset, Reference: ref,
		// client_secret lets the client confirm the PaymentIntent (Stripe.js).
		ClientSecret: stringField(res, "client_secret"),
	}, nil
}

func (s *Stripe) CreatePayout(amount int64, asset, dest, ref string) (Payout, error) {
	form := url.Values{}
	form.Set("amount", strconv.FormatInt(amount, 10))
	form.Set("currency", currencyOf(asset))
	form.Set("metadata[destination]", dest)
	form.Set("metadata[reference]", ref)
	res, err := s.post("/v1/payouts", form)
	if err != nil {
		return Payout{}, err
	}
	return Payout{ExternalID: stringField(res, "id"), Status: StatusSucceeded}, nil
}

func (s *Stripe) post(path string, form url.Values) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodPost, s.cfg.BaseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, newError(ErrPSPUnavailable, err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.cfg.HTTP.Do(req)
	if err != nil {
		return nil, newError(ErrPSPUnavailable, err.Error())
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, newError(ErrPSPUnavailable, fmt.Sprintf("stripe %s -> %d: %s", path, resp.StatusCode, string(body)))
	}
	var out map[string]interface{}
	_ = json.Unmarshal(body, &out)
	return out, nil
}

func stringField(m map[string]interface{}, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

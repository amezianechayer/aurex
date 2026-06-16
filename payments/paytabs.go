package payments

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PayTabsConfig configures the PayTabs (MENA) connector. BaseURL and HTTP are
// injectable for testing against an httptest server.
type PayTabsConfig struct {
	ServerKey string
	ProfileID string
	BaseURL   string
	HTTP      *http.Client
}

type PayTabs struct{ cfg PayTabsConfig }

func NewPayTabs(cfg PayTabsConfig) *PayTabs {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://secure.paytabs.com"
	}
	if cfg.HTTP == nil {
		cfg.HTTP = http.DefaultClient
	}
	return &PayTabs{cfg: cfg}
}

func (p *PayTabs) Name() string { return "paytabs" }

// HandleWebhook verifies the PayTabs `signature` header (HMAC-SHA256 hex of the
// raw payload with the server key) and normalizes the callback. wallet_id and
// asset travel in user_defined (udf1, udf2). Amount is parsed from the decimal
// cart_amount into int64 minor units (2-decimal currencies).
func (p *PayTabs) HandleWebhook(payload []byte, signature string) (Event, error) {
	if err := p.verifySignature(payload, signature); err != nil {
		return Event{}, err
	}
	var cb struct {
		TranRef      string `json:"tran_ref"`
		CartID       string `json:"cart_id"`
		CartAmount   string `json:"cart_amount"`
		CartCurrency string `json:"cart_currency"`
		UserDefined  struct {
			UDF1 string `json:"udf1"`
			UDF2 string `json:"udf2"`
		} `json:"user_defined"`
		PaymentResult struct {
			ResponseStatus string `json:"response_status"`
		} `json:"payment_result"`
	}
	if err := json.Unmarshal(payload, &cb); err != nil {
		return Event{}, newError(ErrInvalidPayload, "invalid paytabs callback json")
	}
	asset := cb.UserDefined.UDF2
	if asset == "" {
		asset = strings.ToUpper(cb.CartCurrency) + ".2"
	}
	out := Event{
		Type: EventUnhandled, PSP: p.Name(), ExternalID: cb.TranRef,
		WalletID: cb.UserDefined.UDF1, Reference: cb.CartID,
		Asset: asset, Amount: parseMinor(cb.CartAmount, 2),
	}
	switch strings.ToUpper(cb.PaymentResult.ResponseStatus) {
	case "A": // Authorised
		out.Type = EventPaymentSucceeded
	case "D", "E", "C": // Declined / Error / Cancelled
		out.Type = EventPaymentFailed
	}
	return out, nil
}

func (p *PayTabs) verifySignature(payload []byte, signature string) error {
	if p.cfg.ServerKey == "" {
		return newError(ErrInvalidSignature, "no server key configured")
	}
	if signature == "" {
		return newError(ErrInvalidSignature, "missing paytabs signature")
	}
	mac := hmac.New(sha256.New, []byte(p.cfg.ServerKey))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return newError(ErrInvalidSignature, "paytabs signature mismatch")
	}
	return nil
}

// SignWebhook builds a valid PayTabs signature for the payload (for tests).
func (p *PayTabs) SignWebhook(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(p.cfg.ServerKey))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func (p *PayTabs) CreatePayment(amount int64, asset, ref string) (Payment, error) {
	body := map[string]interface{}{
		"profile_id":    p.cfg.ProfileID,
		"tran_type":     "sale",
		"tran_class":    "ecom",
		"cart_id":       ref,
		"cart_currency": strings.ToUpper(currencyOf(asset)),
		"cart_amount":   formatMinor(amount, 2),
		"cart_description": "wallet topup",
	}
	res, err := p.post("/payment/request", body)
	if err != nil {
		return Payment{}, err
	}
	return Payment{ExternalID: stringField(res, "tran_ref"), Status: StatusPending, Amount: amount, Asset: asset, Reference: ref}, nil
}

func (p *PayTabs) CreatePayout(amount int64, asset, dest, ref string) (Payout, error) {
	body := map[string]interface{}{
		"profile_id":    p.cfg.ProfileID,
		"tran_type":     "payout",
		"cart_id":       ref,
		"cart_currency": strings.ToUpper(currencyOf(asset)),
		"cart_amount":   formatMinor(amount, 2),
		"beneficiary":   dest,
	}
	res, err := p.post("/payment/request", body)
	if err != nil {
		return Payout{}, err
	}
	return Payout{ExternalID: stringField(res, "tran_ref"), Status: StatusSucceeded}, nil
}

func (p *PayTabs) post(path string, body map[string]interface{}) (map[string]interface{}, error) {
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, p.cfg.BaseURL+path, bytes.NewReader(raw))
	if err != nil {
		return nil, newError(ErrPSPUnavailable, err.Error())
	}
	req.Header.Set("authorization", p.cfg.ServerKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.cfg.HTTP.Do(req)
	if err != nil {
		return nil, newError(ErrPSPUnavailable, err.Error())
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, newError(ErrPSPUnavailable, fmt.Sprintf("paytabs %s -> %d: %s", path, resp.StatusCode, string(b)))
	}
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	return out, nil
}

// parseMinor turns a decimal string ("10.50") into int64 minor units given a
// fixed number of decimals, using integer arithmetic only (never float).
func parseMinor(s string, decimals int) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	intPart, fracPart := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, fracPart = s[:i], s[i+1:]
	}
	if len(fracPart) > decimals {
		fracPart = fracPart[:decimals]
	}
	for len(fracPart) < decimals {
		fracPart += "0"
	}
	var n int64
	for _, c := range intPart + fracPart {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int64(c-'0')
	}
	if neg {
		n = -n
	}
	return n
}

// formatMinor is the inverse of parseMinor for request bodies.
func formatMinor(amount int64, decimals int) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}
	digits := fmt.Sprintf("%0*d", decimals+1, amount)
	out := digits[:len(digits)-decimals] + "." + digits[len(digits)-decimals:]
	if neg {
		out = "-" + out
	}
	return out
}
